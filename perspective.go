/*
 * Copyright 2019 Aletheia Ware LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package perspectivego

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

func ReadWorldFile(path string) (*World, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	buffer, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	size, s := proto.DecodeVarint(buffer[:])
	if s <= 0 {
		return nil, errors.New("Could not read size")
	}
	world := &World{}
	if err = proto.Unmarshal(buffer[s:s+int(size)], world); err != nil {
		return nil, err
	}
	return world, nil
}

func WriteWorldFile(path string, world *World) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()
	return WriteWorld(file, world)
}

func WriteWorld(writer io.Writer, world *World) error {
	size := uint64(proto.Size(world))

	data, err := proto.Marshal(world)
	if err != nil {
		return err
	}
	if _, err := writer.Write(proto.EncodeVarint(size)); err != nil {
		return err
	}
	if _, err := writer.Write(data); err != nil {
		return err
	}
	return nil
}

func ReadPuzzle(reader io.Reader) (*Puzzle, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	var outline *Outline
	var block []*Block
	var goal []*Goal
	var portal []*Portal
	var sphere []*Sphere
	var description string
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, ":")
		switch parts[0] {
		case "description":
			description = parts[1]
		case "outline":
			outline = &Outline{
				Type:   parts[1],
				Colour: parts[2],
			}
		case "block":
			block = append(block, &Block{
				Name:     parts[1],
				Type:     parts[2],
				Colour:   parts[3],
				Location: ParseLocation(parts[4]),
			})
		case "goal":
			goal = append(goal, &Goal{
				Name:     parts[1],
				Type:     parts[2],
				Colour:   parts[3],
				Location: ParseLocation(parts[4]),
			})
		case "portal":
			portal = append(portal, &Portal{
				Name:     parts[1],
				Type:     parts[2],
				Colour:   parts[3],
				Location: ParseLocation(parts[4]),
				Link:     ParseLocation(parts[5]),
			})
		case "sphere":
			sphere = append(sphere, &Sphere{
				Name:     parts[1],
				Type:     parts[2],
				Colour:   parts[3],
				Location: ParseLocation(parts[4]),
			})
		}
	}
	return &Puzzle{
		Outline:     outline,
		Block:       block,
		Goal:        goal,
		Portal:      portal,
		Sphere:      sphere,
		Description: description,
	}, nil
}

func ParseLocation(s string) *Location {
	parts := strings.Split(s, ",")
	w := 0
	x := 0
	y := 0
	z := 0
	switch len(parts) {
	case 4:
		w = StringToInt(parts[0])
		x = StringToInt(parts[1])
		y = StringToInt(parts[2])
		z = StringToInt(parts[3])
	case 3:
		x = StringToInt(parts[0])
		y = StringToInt(parts[1])
		z = StringToInt(parts[2])
	case 2:
		x = StringToInt(parts[0])
		y = StringToInt(parts[1])
	case 1:
		x = StringToInt(parts[0])
	case 0:
		fallthrough
	default:
		log.Fatal("Could not parse location", s)
	}
	return &Location{
		W: int32(w),
		X: int32(x),
		Y: int32(y),
		Z: int32(z),
	}
}

func StringToInt(s string) int {
	index, err := strconv.Atoi(s)
	if err != nil {
		log.Fatal(err)
	}
	return index
}

func UnparseLocation(l *Location) string {
	if l.W != 0 {
		return strconv.Itoa(int(l.W)) + "," + strconv.Itoa(int(l.X)) + "," + strconv.Itoa(int(l.Y)) + "," + strconv.Itoa(int(l.Z))
	}
	return strconv.Itoa(int(l.X)) + "," + strconv.Itoa(int(l.Y)) + "," + strconv.Itoa(int(l.Z))
}

func WritePuzzleFile(path string, puzzle *Puzzle) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()
	return WritePuzzle(file, puzzle)
}

func WritePuzzle(writer io.Writer, puzzle *Puzzle) error {
	if puzzle.Outline != nil {
		fmt.Fprintln(writer, "outline:"+puzzle.Outline.Type+":"+puzzle.Outline.Colour)
	}
	if puzzle.Description != "" {
		fmt.Fprintln(writer, "description:"+puzzle.Description)
	}
	for _, b := range puzzle.Block {
		fmt.Fprintln(writer, "block:"+b.Name+":"+b.Type+":"+b.Colour+":"+UnparseLocation(b.Location))
	}
	for _, g := range puzzle.Goal {
		fmt.Fprintln(writer, "goal:"+g.Name+":"+g.Type+":"+g.Colour+":"+UnparseLocation(g.Location))
	}
	for _, p := range puzzle.Portal {
		fmt.Fprintln(writer, "portal:"+p.Name+":"+p.Type+":"+p.Colour+":"+UnparseLocation(p.Location)+":"+UnparseLocation(p.Link))
	}
	for _, s := range puzzle.Sphere {
		fmt.Fprintln(writer, "sphere:"+s.Name+":"+s.Type+":"+s.Colour+":"+UnparseLocation(s.Location))
	}
	return nil
}
