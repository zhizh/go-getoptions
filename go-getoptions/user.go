// This file is part of go-getoptions.
//
// Copyright (C) 2015-2021  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package getoptions

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/DavidGamba/go-getoptions/option"
)

var Logger = log.New(ioutil.Discard, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)

// exitFn - This variable allows to test os.Exit calls
var exitFn = os.Exit

// completionWriter - Writer where the completion results will be written to.
// Set as a variable to allow for easy testing.
var completionWriter io.Writer = os.Stdout

type GetOpt struct {
	programTree *programTree
}

// Mode - Operation mode for short options
type Mode int

// Operation modes
const (
	Normal Mode = iota
	Bundling
	SingleDash
)

// UnknownMode - Unknown option mode
type UnknownMode int

// Unknown option modes
const (
	Fail UnknownMode = iota
	Warn
	Pass
)

// CommandFn - Function signature for commands
type CommandFn func(context.Context, *GetOpt, []string) error

type ModifyFn func(string)

func New() *GetOpt {
	gopt := &GetOpt{}
	gopt.programTree = &programTree{
		Type:          argTypeProgname,
		Name:          os.Args[0],
		ChildCommands: map[string]*programTree{},
		ChildOptions:  map[string]*option.Option{},
		Level:         0,
	}
	return gopt
}

func (gopt *GetOpt) NewCommand(name string, description string) *GetOpt {
	cmd := &GetOpt{}
	command := &programTree{
		Type:          argTypeCommand,
		Name:          name,
		ChildCommands: map[string]*programTree{},
		ChildOptions:  map[string]*option.Option{},
		Parent:        gopt.programTree,
		Level:         gopt.programTree.Level + 1,
	}

	// Copy option definitions from parent to child
	for k, v := range gopt.programTree.ChildOptions {
		// The option parent doesn't match properly here.
		// I should in a way create a copy of the option but I still want a pointer to the data.

		// c := v.Copy() // copy that maintains a pointer to the underlying data
		// c.SetParent(command)

		// TODO: This is doing an overwrite, ensure it doesn't exist
		// command.ChildOptions[k] = c
		command.ChildOptions[k] = v
	}
	cmd.programTree = command
	gopt.programTree.ChildCommands[name] = command
	return cmd
}

func (gopt *GetOpt) String(name, def string, fns ...ModifyFn) *string {
	gopt.StringVar(&def, name, def, fns...)
	return &def
}

func (gopt *GetOpt) StringVar(p *string, name, def string, fns ...ModifyFn) {
	n := option.New(name, option.StringType, &def)
	gopt.programTree.ChildOptions[name] = n
}

func (gopt *GetOpt) Parse(args []string) ([]string, error) {
	compLine := os.Getenv("COMP_LINE")
	if compLine != "" {
		Logger.Printf("COMP_LINE: %s\n", compLine)
		_, completions, err := parseCLIArgs(true, gopt.programTree, args, Normal)
		if err != nil {
			return nil, err
		}
		fmt.Fprintln(completionWriter, strings.Join(*completions, "\n"))
		exitFn(124) // programmable completion restarts from the beginning, with an attempt to find a new compspec for that command.
	}
	return nil, nil
}
