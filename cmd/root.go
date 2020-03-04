/*Package cmd contains all commands
Copyright © 2019 Mikael Fridh <frimik@gmail.com>

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
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"encoding/json"
	"io/ioutil"

	log "github.com/sirupsen/logrus"

	"sort"

	"github.com/spf13/viper"

	serversets "github.com/frimik/go.serversets"
)

var (
	zookeeperHosts     []string // Zookeeper hosts
	auroraRole         string
	auroraEnv          string
	auroraJob          string
	mcRouterConfigFile string
	verbose            bool
	debug              bool
	mcrouterExe        string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mczoorouter",
	Short: "Configure mcrouter based on a Zookeeper Finagle ServerSet",
	Long:  ` `,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {},
}

func byShards(el []serversets.Entity) []string {
	shards := make([]string, 0, len(el))

	sort.Sort(EntityByShard(el))
	for _, server := range el {
		shards = append(shards, fmt.Sprintf("%s:%d", server.ServiceEndpoint.Host, server.ServiceEndpoint.Port))
	}

	return shards
}

// '{"pools":{"A":{"servers":["cache:11211"]}},"route":"PoolRoute|A"}' -p 5000

type Configuration struct {
	Pools Pools  `json:"pools"`
	Route string `json:"route"`
}

type Pools struct {
	Pool Pool `json:"A"`
}

type Pool struct {
	Servers []string `json:"servers"`
}

// EntityByShard implements sort.Interface based on the Shard field.
type EntityByShard []serversets.Entity

func (a EntityByShard) Len() int           { return len(a) }
func (a EntityByShard) Less(i, j int) bool { return a[i].Shard < a[j].Shard }
func (a EntityByShard) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func writeConfig(watch *serversets.Watch) error {

	var jsonData []byte
	_servers := byShards(watch.EndpointEntities())
	mcConfig := Configuration{
		Pools{
			Pool{
				Servers: _servers,
			},
		},
		"PoolRoute|A",
	}
	jsonData, err := json.MarshalIndent(&mcConfig, "", "    ")

	if err != nil {
		log.Println(err)
	}
	fmt.Println(string(jsonData))

	_ = ioutil.WriteFile("mcrouter.json", jsonData, 0644)

	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	serversets.BaseDirectory = "/aurora/jobs"

	log.Infof("Using ServerSet path: %s/%s/%s/%s", serversets.BaseDirectory, auroraRole, auroraEnv, auroraJob)
	serverSet := serversets.New(
		auroraRole, auroraEnv, auroraJob, zookeeperHosts,
	)

	watch, err := serverSet.Watch()
	if err != nil {
		// probably something wrong with connecting to Zookeeper
		panic(err)
	}

	// initial configuration write
	writeConfig(watch)

	for {
		<-watch.Event()
		writeConfig(watch)
	}

}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.Flags().BoolVarP(&debug, "debug", "d", false, "debug output")

	rootCmd.Flags().StringSliceVar(&zookeeperHosts, "zookeeper", []string{}, "zookeeper hosts")
	rootCmd.MarkFlagRequired("zookeeper")
	rootCmd.Flags().StringVar(&auroraRole, "role", "", "Memcache Aurora Role")
	rootCmd.MarkFlagRequired("role")
	rootCmd.Flags().StringVar(&auroraEnv, "env", "", "Memcache Aurora Environment")
	rootCmd.MarkFlagRequired("env")
	rootCmd.Flags().StringVar(&auroraJob, "job", "", "Memcache Aurora Job")
	rootCmd.MarkFlagRequired("job")

	rootCmd.Flags().StringVar(&mcRouterConfigFile, "mcrouter-config-file", "", "mcRouter configuration file.")
	rootCmd.MarkFlagRequired("mcrouter-config-file")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetEnvPrefix("mczoorouter")
	viper.AutomaticEnv() // read in environment variables that match
}
