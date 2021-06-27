package getoptions

import (
	"errors"
	"reflect"
	"testing"
)

func checkError(t *testing.T, got, expected error) {
	t.Helper()
	if (got == nil && expected != nil) || (got != nil && expected == nil) || (got != nil && expected != nil && !errors.Is(got, expected)) {
		t.Errorf("wrong error received: got = '%#v', want '%#v'", got, expected)
	}
}

func TestParseCLIArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		mode     Mode
		expected *programTree
		err      error
	}{

		{"empty", nil, Normal, setupOpt().programTree, nil},

		{"empty", []string{}, Normal, setupOpt().programTree, nil},

		{"text", []string{"txt"}, Normal, func() *programTree {
			tree := setupOpt().programTree
			value := "txt"
			tree.ChildText = append(tree.ChildText, &value)
			return tree
		}(), nil},

		{"command", []string{"cmd1"}, Normal, func() *programTree {
			n, err := getNode(setupOpt().programTree, "cmd1")
			if err != nil {
				panic(err)
			}
			return n
		}(), nil},

		{"text to command", []string{"cmd1", "txt"}, Normal, func() *programTree {
			n, err := getNode(setupOpt().programTree, "cmd1")
			if err != nil {
				panic(err)
			}
			value := "txt"
			n.ChildText = append(n.ChildText, &value)
			return n
		}(), nil},

		{"text to sub command", []string{"cmd1", "sub1cmd1", "txt"}, Normal, func() *programTree {
			n, err := getNode(setupOpt().programTree, "cmd1", "sub1cmd1")
			if err != nil {
				panic(err)
			}
			value := "txt"
			n.ChildText = append(n.ChildText, &value)
			return n
		}(), nil},

		{"option with arg", []string{"--rootopt1=hello"}, Normal, func() *programTree {
			tree := setupOpt().programTree
			opt, ok := tree.ChildOptions["rootopt1"]
			if !ok {
				t.Fatalf("not found")
			}
			opt.Called = true
			opt.UsedAlias = "rootopt1"
			err := opt.Save("hello")
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			return tree
		}(), nil},

		{"option", []string{"--rootopt1", "hello"}, Normal, func() *programTree {
			tree := setupOpt().programTree
			opt, ok := tree.ChildOptions["rootopt1"]
			if !ok {
				t.Fatalf("not found")
			}
			opt.Called = true
			opt.UsedAlias = "rootopt1"
			err := opt.Save("hello")
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			return tree
		}(), nil},

		{"option error missing argument", []string{"--rootopt1"}, Normal, func() *programTree {
			tree := setupOpt().programTree
			opt, ok := tree.ChildOptions["rootopt1"]
			if !ok {
				t.Fatalf("not found")
			}
			opt.Called = true
			opt.UsedAlias = "rootopt1"
			return tree
		}(), ErrorMissingArgument},

		{"terminator", []string{"--", "--opt1"}, Normal, func() *programTree {
			tree := setupOpt().programTree
			value := "--opt1"
			tree.ChildText = append(tree.ChildText, &value)
			return tree
		}(), nil},

		{"lonesome dash", []string{"cmd1", "sub2cmd1", "-"}, Normal, func() *programTree {
			tree := setupOpt().programTree
			n, err := getNode(tree, "cmd1", "sub2cmd1")
			if err != nil {
				t.Fatalf("unexpected error: %s, %s", err, n.Str())
			}
			opt, ok := n.ChildOptions["-"]
			if !ok {
				t.Fatalf("not found: %s", n.Str())
			}
			opt.Called = true
			opt.UsedAlias = "-"
			return n
		}(), nil},

		{"root option to command", []string{"cmd1", "--rootopt1", "hello"}, Normal, func() *programTree {
			tree := setupOpt().programTree
			n, err := getNode(tree, "cmd1")
			if err != nil {
				t.Fatalf("unexpected error: %s, %s", err, n.Str())
			}
			opt, ok := n.ChildOptions["rootopt1"]
			if !ok {
				t.Fatalf("not found: %s", n.Str())
			}
			opt.Called = true
			opt.UsedAlias = "rootopt1"
			err = opt.Save("hello")
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			return n
		}(), nil},

		{"root option to subcommand", []string{"cmd1", "sub2cmd1", "--rootopt1", "hello"}, Normal, func() *programTree {
			tree := setupOpt().programTree
			n, err := getNode(tree, "cmd1", "sub2cmd1")
			if err != nil {
				t.Fatalf("unexpected error: %s, %s", err, n.Str())
			}
			opt, ok := n.ChildOptions["rootopt1"]
			if !ok {
				t.Fatalf("not found: %s", n.Str())
			}
			opt.Called = true
			opt.UsedAlias = "rootopt1"
			err = opt.Save("hello")
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			return n
		}(), nil},

		{"option to subcommand", []string{"cmd1", "sub1cmd1", "--sub1cmd1opt1=hello"}, Normal, func() *programTree {
			tree := setupOpt().programTree
			n, err := getNode(tree, "cmd1", "sub1cmd1")
			if err != nil {
				t.Fatalf("unexpected error: %s, %s", err, n.Str())
			}
			opt, ok := n.ChildOptions["sub1cmd1opt1"]
			if !ok {
				t.Fatalf("not found: %s", n.Str())
			}
			opt.Called = true
			opt.UsedAlias = "sub1cmd1opt1"
			err = opt.Save("hello")
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			return n
		}(), nil},

		{"option to subcommand", []string{"cmd1", "sub1cmd1", "--sub1cmd1opt1", "hello"}, Normal, func() *programTree {
			tree := setupOpt().programTree
			n, err := getNode(tree, "cmd1", "sub1cmd1")
			if err != nil {
				t.Fatalf("unexpected error: %s, %s", err, n.Str())
			}
			opt, ok := n.ChildOptions["sub1cmd1opt1"]
			if !ok {
				t.Fatalf("not found: %s", n.Str())
			}
			opt.Called = true
			opt.UsedAlias = "sub1cmd1opt1"
			err = opt.Save("hello")
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			return n
		}(), nil},

		{"option argument with dash", []string{"cmd1", "sub1cmd1", "--sub1cmd1opt1", "-hello"}, Normal, func() *programTree {
			tree := setupOpt().programTree
			n, err := getNode(tree, "cmd1", "sub1cmd1")
			if err != nil {
				t.Fatalf("unexpected error: %s, %s", err, n.Str())
			}
			opt, ok := n.ChildOptions["sub1cmd1opt1"]
			if !ok {
				t.Fatalf("not found: %s", n.Str())
			}
			opt.Called = true
			opt.UsedAlias = "sub1cmd1opt1"
			return n
		}(), ErrorMissingArgument},

		// {"command", []string{"--opt1", "cmd1", "--cmd1opt1"}, Normal, &programTree{
		// 	Type:   argTypeProgname,
		// 	Name:   os.Args[0],
		// 	option: option{Args: []string{"--opt1", "cmd1", "--cmd1opt1"}},
		// 	Children: []*programTree{
		// 		{
		// 			Type:     argTypeOption,
		// 			Name:     "opt1",
		// 			option:   option{Args: []string{}},
		// 			Children: []*programTree{},
		// 		},
		// 		{
		// 			Type:   argTypeCommand,
		// 			Name:   "cmd1",
		// 			option: option{Args: []string{}},
		// 			Children: []*programTree{
		// 				{
		// 					Type:     argTypeOption,
		// 					Name:     "cmd1opt1",
		// 					option:   option{Args: []string{}},
		// 					Children: []*programTree{},
		// 				},
		// 			},
		// 		},
		// 	},
		// }},
		// {"subcommand", []string{"--opt1", "cmd1", "--cmd1opt1", "sub1cmd1", "--sub1cmd1opt1"}, Normal, &programTree{
		// 	Type:   argTypeProgname,
		// 	Name:   os.Args[0],
		// 	option: option{Args: []string{"--opt1", "cmd1", "--cmd1opt1", "sub1cmd1", "--sub1cmd1opt1"}},
		// 	Children: []*programTree{
		// 		{
		// 			Type:     argTypeOption,
		// 			Name:     "opt1",
		// 			option:   option{Args: []string{}},
		// 			Children: []*programTree{},
		// 		},
		// 		{
		// 			Type:   argTypeCommand,
		// 			Name:   "cmd1",
		// 			option: option{Args: []string{}},
		// 			Children: []*programTree{
		// 				{
		// 					Type:     argTypeOption,
		// 					Name:     "cmd1opt1",
		// 					option:   option{Args: []string{}},
		// 					Children: []*programTree{},
		// 				},
		// 				{
		// 					Type:   argTypeCommand,
		// 					Name:   "sub1cmd1",
		// 					option: option{Args: []string{}},
		// 					Children: []*programTree{
		// 						{
		// 							Type:     argTypeOption,
		// 							Name:     "sub1cmd1opt1",
		// 							option:   option{Args: []string{}},
		// 							Children: []*programTree{},
		// 						},
		// 					},
		// 				},
		// 			},
		// 		},
		// 	},
		// }},
		// {"arg", []string{"hello", "world"}, Normal, &programTree{
		// 	Type:   argTypeProgname,
		// 	Name:   os.Args[0],
		// 	option: option{Args: []string{"hello", "world"}},
		// 	Children: []*programTree{
		// 		{
		// 			Type:     argTypeText,
		// 			Name:     "hello",
		// 			option:   option{Args: []string{}},
		// 			Children: []*programTree{},
		// 		},
		// 		{
		// 			Type:     argTypeText,
		// 			Name:     "world",
		// 			option:   option{Args: []string{}},
		// 			Children: []*programTree{},
		// 		},
		// 	},
		// }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := setupLogging()
			tree := setupOpt().programTree
			argTree, _, err := parseCLIArgs(false, tree, test.args, test.mode)
			checkError(t, err, test.err)
			if !reflect.DeepEqual(test.expected, argTree) {
				t.Errorf("expected tree, got: %s %s\n", SpewToFile(t, test.expected, "expected"), SpewToFile(t, argTree, "got"))
				t.Fatalf("expected tree: \n%s\n got: \n%s\n", test.expected.Str(), argTree.Str())
			}
			if len(buf.String()) > 0 {
				t.Log("\n" + buf.String())
			}
		})
	}
}
