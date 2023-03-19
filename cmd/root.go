/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spf13/cobra"
)

var (
	threads int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "port-scan",
	Short: "A tool for scanning open ports of a specified IP address",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Please specify the port to be scanned")
			return
		}

		if threads <= 0 {
			threads = 2000
		}

		var (
			host  = args[0]
			ports []int
			ch    = make(chan int, threads)
			lock  = sync.Mutex{}
			wg    = sync.WaitGroup{}
			total = int64(0)
		)
		for i := 0; i < threads; i++ {
			go func() {
				for {
					select {
					case port := <-ch:
						conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), time.Second*2)
						lock.Lock()
						if err == nil && conn != nil {
							ports = append(ports, port)
						}
						if atomic.AddInt64(&total, 1)%int64(threads) == 0 {
							fmt.Printf("Scanned %d ports, %d open ports found\n", total, len(ports))
						}
						lock.Unlock()
						wg.Done()
					}
				}
			}()
		}
		for i := 1; i < 65535; i++ {
			wg.Add(1)
			port := i
			ch <- port
		}
		wg.Wait()
		fmt.Printf("open ports: %v\n", ports)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.Use = "port-scan host"
	rootCmd.Flags().IntVar(&threads, "threads", 2000, "Number of threads to use")
}
