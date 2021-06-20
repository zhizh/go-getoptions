package getoptions

import (
	"fmt"
	"sort"
	"strings"

	"github.com/DavidGamba/go-getoptions/option"
	"github.com/DavidGamba/go-getoptions/sliceiterator"
)

type programTree struct {
	Type          argType
	Name          string
	ChildCommands map[string]*programTree
	ChildOptions  map[string]*option.Option
	ChildText     []*string
	Parent        *programTree
	Level         int
	command
}

// Str - not String so it doesn't get called automatically by Spew.
func (n *programTree) Str() string {
	level := n.Level
	if n.Type == argTypeOption {
		if n.Parent != nil {
			level = n.Parent.Level + 1
		}
	}
	padding := func(n int) string {
		return strings.Repeat("  ", n)
	}
	out := padding(level) + fmt.Sprintf("Name: %v, Type: %v", n.Name, n.Type)
	if n.Parent != nil {
		out += fmt.Sprintf(", Parent: %v", n.Parent.Name)
	}
	if len(n.ChildOptions) > 0 {
		var keys []string
		for k := range n.ChildOptions {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		out += ", child options: [\n"
		for _, k := range keys {
			out += padding(level+1) + fmt.Sprintf("Name: %s, Aliases %v\n", n.ChildOptions[k].Name, n.ChildOptions[k].Aliases)
		}
		out += padding(level) + "]"
	} else {
		out += ", child options: []"
	}
	if len(n.ChildCommands) > 0 {
		var keys []string
		for k := range n.ChildCommands {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		out += ", child commands: [\n"
		for _, k := range keys {
			out += n.ChildCommands[k].Str()
		}
		out += padding(level) + "]"
	} else {
		out += ", child commands: []"
	}
	out += "\n"
	return out
}

// Copy - Returns a copy of programTree that maintains a pointer to the underlying data
func (n *programTree) Copy() *programTree {
	// a := *n
	// c := &a
	parent := *n.Parent
	c := &programTree{
		Type:          n.Type,
		Name:          n.Name,
		ChildCommands: n.ChildCommands,
		ChildOptions:  n.ChildOptions,
		ChildText:     n.ChildText,
		Parent:        &parent,
	}
	return c
}

func (n *programTree) SetParent(p *programTree) *programTree {
	n.Parent = p
	return n
}

func getNode(tree *programTree, element ...string) (*programTree, error) {
	if len(element) == 0 {
		return tree, nil
	}
	if child, ok := tree.ChildCommands[element[0]]; ok {
		return getNode(child, element[1:]...)
	}
	return tree, fmt.Errorf("not found")
}

type argType int

const (
	argTypeProgname   argType = iota // The root node type
	argTypeCommand                   // The node type used for commands and subcommands
	argTypeOption                    // The node type used for options
	argTypeText                      // The node type used for regular cli arguments
	argTypeTerminator                // --
)

// command - Fields that only make sense for a command
type command struct {
	CommandFn CommandFn
}

// TODO: Make this a method of tree so we can add parent information.
// Maybe not a good idea? Would it complicate testing?
func newCLIOption(parent *programTree, name string, args ...string) *option.Option {
	data := []string{}
	data = append(data, args...)
	arg := option.New(name, option.StringRepeatType, &data)
	arg.Unknown = true
	return arg
}

type completions *[]string

// parseCLIArgs - Given the root node tree and the cli args it returns a populated tree of the node that was called.
// For example, if a command is called, then the returned node is that of the command with the options that were set updated with their values.
// Additionally, when in completion mode, it returns the possible completions
func parseCLIArgs(completionMode bool, tree *programTree, args []string, mode Mode) (*programTree, completions, error) {
	// Design: This function could return an array or CLIargs as a parse result
	// or I could do one level up and have a root CLIarg type with the name of
	// the program.  Having the root level might be helpful with help generation.

	// The current implementation expects os.Args[1:] as an argument so this
	// can't expect to receive the os.Args[0] as the first argument.

	// CLI arguments are split by spaces by the shell and passed as individual
	// strings.  In most cases, a cli argument (one string) represents one option
	// or one argument, however, in the case of bundling mode a single string can
	// represent multiple options.

	// Ensure consistent response for empty and nil slices
	if args == nil {
		args = []string{}
	}

	currentProgramNode := tree

	iterator := sliceiterator.New(&args)

ARGS_LOOP:
	for iterator.Next() {

		// We only generate completions when we reached the end of the provided args
		if completionMode && iterator.IsLast() {
			// TODO: check what was the behaviour when you have a space and hit the tab completion.

			// TODO: Handle completions
			// We check to see if this is the last arg and act on that one.
			if iterator.Value() == "-" || iterator.Value() == "--" {
				// Provide option completions
			}
			if strings.HasPrefix(iterator.Value(), "-") {
				// Provide option completions
			}
			// Iterate over commands and check prefix to see if we offer command completion

			// Provide other kinds of completions, like file completions.
		}

		// handle terminator
		if iterator.Value() == "--" {
			for iterator.Next() {
				value := iterator.Value()
				currentProgramNode.ChildText = append(currentProgramNode.ChildText, &value)
			}
			break
		}

		// Handle lonesome dash
		if iterator.Value() == "-" {
			for _, v := range currentProgramNode.ChildOptions {
				// handle full option match, this allows to have - defined as an alias
				if _, ok := stringSliceIndex(v.Aliases, "-"); ok {
					v.Called = true
					v.UsedAlias = "-"
					continue ARGS_LOOP
				}
			}
			opt := newCLIOption(currentProgramNode, "-")
			currentProgramNode.ChildOptions["-"] = opt
			continue ARGS_LOOP
		}

		// TODO: Handle unknonw option.
		// It basically needs to be copied down to the command every time we find a command and it has to be validated against aliases and option name.
		// If we were to check on require order and other modes without doing that work, passing --help after passing an unknown option would return an unknown option error and it would be annoying to the user.

		// TODO: Handle case where option has an argument
		// check for option

		// isOption should check if a cli argument starts with -.
		// If it does, we validate that it matches an option.
		// If it does we update the option with the values that might have been provided on the CLI.
		//
		// We almost need to build a separate option tree which allows unknown options and then update the main tree when we are done parsing cli args.
		//
		// Currently go-getoptions has no knowledge of command options at the
		// parents so it marks them as an unkonw option that needs to be used at a
		// different level. It is as if it was ignoring getoptions.Pass.
		if optPair, is := isOption(iterator.Value(), mode, false); is {
			// iterate over the possible cli args and try matching against expectations
			for _, p := range optPair {
				matches := 0
				for _, c := range currentProgramNode.ChildOptions {
					// handle full option match
					// TODO: handle partial matches
					if _, ok := stringSliceIndex(c.Aliases, p.Option); ok {
						c.Called = true
						c.UsedAlias = p.Option
						err := c.Save(p.Args...)
						if err != nil {
							// TODO: This shouldn't happen, figure out what to do about it.
							Logger.Println("this shouldn't happen")
						}
						matches++
						// TODO: Handle option having a minimum bigger than 1
					}
				}
				if matches > 1 {
					// TODO: handle ambiguous option call error
					continue
				}
				if matches == 0 {
					// TODO: This is a new option, add it as a children and mark it as unknown
					// TODO: This shouldn't append new children but update existing ones and isOption needs to be able to check if the option expects a follow up argument.
					// Check min, check max and keep ingesting until something starts with `-` or matches a command.

					opt := newCLIOption(currentProgramNode, p.Option, p.Args...)
					currentProgramNode.ChildOptions[p.Option] = opt
				}
			}
			continue
		}

		// When handling options out of order, iterate over all possible options for all the children and set them if they match.
		// That means that the option has to match the alias and aliases need to be non ambiguous with the parent.
		// partial options can only be applied if they match a single possible option in the tree.
		// Since at the end we return the programTree node, we will only care about handling the options at one single level.

		// handle commands and subcommands
		for k, v := range currentProgramNode.ChildCommands {
			if k == iterator.Value() {
				currentProgramNode = v
				continue ARGS_LOOP
			}
		}

		// handle text
		value := iterator.Value()
		currentProgramNode.ChildText = append(currentProgramNode.ChildText, &value)
	}

	// TODO: Before returning the current node, parse EnvVars and update the values.

	// TODO: After being done parsing everything validate for errors
	// Errors can be unknown options, options without values, etc

	return currentProgramNode, &[]string{}, nil
}

// TODO:
// suggestCompletions -
func suggestCompletions(tree *programTree, args []string, mode Mode) {}
