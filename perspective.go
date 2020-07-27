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
	var sky []*Sky
	var block []*Block
	var goal []*Goal
	var portal []*Portal
	var sphere []*Sphere
	var description string
	var target int
	var scenery []*Scenery
	var dialog []*Dialog
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, ":")
		switch parts[0] {
		case "description":
			description = parts[1]
		case "target":
			target = StringToInt(parts[1])
		case "outline":
			outline = &Outline{
				Mesh:     parts[1],
				Colour:   parts[2],
				Texture:  parts[3],
				Material: parts[4],
				Shader:   parts[5],
			}
		case "sky":
			sky = append(sky, &Sky{
				Name:     parts[1],
				Mesh:     parts[2],
				Colour:   parts[3],
				Texture:  parts[4],
				Material: parts[5],
				Shader:   parts[6],
			})
		case "block":
			block = append(block, &Block{
				Name:     parts[1],
				Mesh:     parts[2],
				Colour:   parts[3],
				Location: StringToLocation(parts[4]),
				Texture:  parts[5],
				Material: parts[6],
				Shader:   parts[7],
			})
		case "goal":
			goal = append(goal, &Goal{
				Name:     parts[1],
				Mesh:     parts[2],
				Colour:   parts[3],
				Location: StringToLocation(parts[4]),
				Texture:  parts[5],
				Material: parts[6],
				Shader:   parts[7],
			})
		case "portal":
			portal = append(portal, &Portal{
				Name:     parts[1],
				Mesh:     parts[2],
				Colour:   parts[3],
				Location: StringToLocation(parts[4]),
				Link:     StringToLocation(parts[5]),
				Texture:  parts[6],
				Material: parts[7],
				Shader:   parts[8],
			})
		case "sphere":
			sphere = append(sphere, &Sphere{
				Name:     parts[1],
				Mesh:     parts[2],
				Colour:   parts[3],
				Location: StringToLocation(parts[4]),
				Texture:  parts[5],
				Material: parts[6],
				Shader:   parts[7],
			})
		case "scenery":
			scenery = append(scenery, &Scenery{
				Name:     parts[1],
				Mesh:     parts[2],
				Colour:   parts[3],
				Location: StringToLocation(parts[4]),
				Texture:  parts[5],
				Material: parts[6],
				Shader:   parts[7],
			})
		case "dialog":
			dialog = append(dialog, &Dialog{
				Name:             parts[1],
				Type:             parts[2],
				BackgroundColour: parts[3],
				ForegroundColour: parts[4],
				Author:           parts[5],
				Content:          parts[6],
				Location:         StringToLocation(parts[7]),
				Element:          strings.Split(parts[8], ","),
			})
		default:
			log.Println("Unrecognized line:", line)
		}
	}
	return &Puzzle{
		Outline:     outline,
		Sky:         sky,
		Block:       block,
		Goal:        goal,
		Portal:      portal,
		Sphere:      sphere,
		Description: description,
		Target:      uint32(target),
		Scenery:     scenery,
		Dialog:      dialog,
	}, nil
}

func StringToLocation(s string) *Location {
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

func LocationToString(l *Location) string {
	if l.W == 0 {
		return strconv.Itoa(int(l.X)) + "," + strconv.Itoa(int(l.Y)) + "," + strconv.Itoa(int(l.Z))
	}
	return strconv.Itoa(int(l.W)) + "," + strconv.Itoa(int(l.X)) + "," + strconv.Itoa(int(l.Y)) + "," + strconv.Itoa(int(l.Z))
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
	fmt.Fprintln(writer, "target:"+fmt.Sprint(puzzle.Target))
	if puzzle.Outline != nil {
		fmt.Fprintln(writer, "outline:"+puzzle.Outline.Mesh+":"+puzzle.Outline.Colour+":"+puzzle.Outline.Texture+":"+puzzle.Outline.Material+":"+puzzle.Outline.Shader)
	}
	for _, s := range puzzle.Sky {
		fmt.Fprintln(writer, "sky:"+s.Name+":"+s.Mesh+":"+s.Colour+":"+s.Texture+":"+s.Material+":"+s.Shader)
	}
	if puzzle.Description != "" {
		fmt.Fprintln(writer, "description:"+puzzle.Description)
	}
	for _, b := range puzzle.Block {
		fmt.Fprintln(writer, "block:"+b.Name+":"+b.Mesh+":"+b.Colour+":"+LocationToString(b.Location)+":"+b.Texture+":"+b.Material+":"+b.Shader)
	}
	for _, g := range puzzle.Goal {
		fmt.Fprintln(writer, "goal:"+g.Name+":"+g.Mesh+":"+g.Colour+":"+LocationToString(g.Location)+":"+g.Texture+":"+g.Material+":"+g.Shader)
	}
	for _, p := range puzzle.Portal {
		fmt.Fprintln(writer, "portal:"+p.Name+":"+p.Mesh+":"+p.Colour+":"+LocationToString(p.Location)+":"+LocationToString(p.Link)+":"+p.Texture+":"+p.Material+":"+p.Shader)
	}
	for _, s := range puzzle.Sphere {
		fmt.Fprintln(writer, "sphere:"+s.Name+":"+s.Mesh+":"+s.Colour+":"+LocationToString(s.Location)+":"+s.Texture+":"+s.Material+":"+s.Shader)
	}
	for _, s := range puzzle.Scenery {
		fmt.Fprintln(writer, "scenery:"+s.Name+":"+s.Mesh+":"+s.Colour+":"+LocationToString(s.Location)+":"+s.Texture+":"+s.Material+":"+s.Shader)
	}
	for _, d := range puzzle.Dialog {
		fmt.Fprintln(writer, "dialog:"+d.Name+":"+d.Type+":"+d.BackgroundColour+":"+d.ForegroundColour+":"+d.Author+":"+d.Content+":"+LocationToString(d.Location)+":"+strings.Join(d.Element, ","))
	}
	return nil
}
