/*
Copyright © 2022 treilik@posteo.de

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/treilik/walder"
	"github.com/treilik/walder/lib"
)

const (
	bindingName = "binding-file"
	graphDir    = "graph-dir"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "walder",
	Short: "A local multi-dimensional abstract graph editor",
	Long: `A graph editor which navigates locally,
facilitates a way to switch between different graph (-types) and
is agnostic to the graph type implementation`,
	Run: func(cmd *cobra.Command, args []string) {
		err := viper.ReadInConfig()
		if err != nil {
			fmt.Printf("There was a error while reading the config:\n%s\n", err)
			os.Exit(1)
		}
		path := viper.GetString(bindingName)
		wald := lib.NewWalder()
		wald.DimensionRegister(adapters...)

		dirPath := viper.GetString(graphDir)
		if dirPath != "" {
			entrys, err := os.ReadDir(dirPath)
			if err == nil {
				var opener []walder.OpenReader
				for _, a := range adapters {
					if o, ok := a.(walder.OpenReader); ok {
						opener = append(opener, o)
					}
				}

				var graphs []walder.Graph
				for _, f := range entrys {
					if !f.Type().IsRegular() {
						continue
					}
					for _, o := range opener {
						if !strings.Contains(strings.ToLower(f.Name()), strings.ToLower(o.String())) {
							continue
						}
						content, err := os.ReadFile(filepath.Join(dirPath, f.Name()))
						if err != nil {
							continue
						}
						g, err := o.Open(bytes.NewBuffer(content))
						if err != nil {
							continue
						}
						graphs = append(graphs, g)
					}
				}
				wald.GraphRegister(graphs...)
			}
		}

		if path == "" {
			fmt.Printf("please set the key binding-file path")
			os.Exit(1)
		}

		if path != "" {
			buf, err := ioutil.ReadFile(path)
			if err != nil {
				fmt.Printf("There was a error while reading the keybindings from file '%s':\n%s\n", path, err)
				os.Exit(1)
			}
			reader := bytes.NewReader(buf)

			err = wald.ReadBindings(reader)
			if err != nil {
				fmt.Printf("There was a error with setting the keybindings:\n%s\n", err)
				os.Exit(1)
			}
		}
		p := tea.NewProgram(wald)
		go func() {
			for true { // TODO close
				p.Send(<-wald.Syncer)
			}
		}()
		p.EnterAltScreen()
		defer p.ExitAltScreen()
		if err := p.Start(); err != nil {
			fmt.Printf("There was a error while starting the program:\n%s\n", err)
			os.Exit(1)
		}

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.walder.yaml)")

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().String(bindingName, "", "path to keybindings file")
	viper.BindPFlag(bindingName, rootCmd.Flags().Lookup(bindingName))
	viper.SetConfigName(".walder")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME")
}
