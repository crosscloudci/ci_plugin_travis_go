/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

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
	"os"

	"context"
	"encoding/json"
	"github.com/davecgh/go-spew/spew"
	"github.com/shuheiktgw/go-travis"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strings"
)

// "https://travis-ci.org/crosscloudci/testproj/builds/572521581"
type Job struct {
	Id   uint
	Href string `json:"@href"`
}

type Commit struct {
	Message string
	Ref     string
	Sha     string
}

type Repository struct {
	Name string
	Slug string
}

type Build struct {
	Id         uint
	State      string
	Commit     Commit
	Repository Repository
	Href       string `json:"@href"`
}

type CliResponse struct {
	JobUrl          string
	BuildUrl        string
	BuildStatus     string
	OptionalMessage string
}

func (c *CliResponse) output() (output string) {
	//TODO if -q parameter don't add header
	fmt.Printf("status\tbuild_url\n")
	fmt.Printf("%v\t%v \n", c.BuildStatus, c.BuildUrl)
	// fmt.Printf("{'build_url': '%v', 'status': '%v'}", c.BuildUrl, c.BuildStatus)
	return
}

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:              "status",
	TraverseChildren: true,
	Short:            "This command retrieves the status of a travis-ci project build",
	Long:             `This command takes a project name, commit ref, or tag and return success, failure, or running.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := travis.NewClient(travis.ApiOrgUrl, os.Getenv("TRAVIS_API_KEY"))
		opt := &travis.BuildsByRepoOption{Limit: 50, Include: []string{"build.commit", "build.branch", "build.repository", "build.jobs"}}
		build, _, err := client.Builds.ListByRepoSlug(context.Background(), viper.GetString("project"), opt)

		if err != nil {
			fmt.Println("Build.Find returned error: ", err)
			os.Exit(1)
		}

		// var retrieveJob = func(build_id uint) (job Job) {
		// 	job = Job{}
		// 	retrieved_job, _, job_err := client.Jobs.ListByBuild(context.Background(), build_id)
		// 	if job_err != nil {
		// 		fmt.Println("Job.Find returned error: ", err)
		// 		os.Exit(1)
		// 	}
		// 	job_json, _ := json.Marshal(retrieved_job[0])
		// 	err := json.Unmarshal(job_json, &job)
		// 	if err != nil {
		// 		panic(err)
		// 		os.Exit(1)
		// 	}
		// 	if viper.GetBool("verbose") {
		// 		spew.Dump("travis job", job)
		// 	}
		// 	return
		// }

		var retrieveBuildStatus = func(b []*travis.Build) (cli_response CliResponse) {
			cli_response = CliResponse{}
			travis_build := Build{}
			// job := Job{}
			var returned_build_status string
			var returned_build_url string
			for _, b := range b {
				build_json, _ := json.Marshal(b)
				err := json.Unmarshal(build_json, &travis_build)
				if err != nil {
					panic(err)
					os.Exit(1)
				}

				arg_commit := viper.GetString("commit")
				if travis_build.Commit.Sha[:6] == arg_commit[:6] {
					// job = retrieveJob(travis_build.Id)
					returned_build_status = travis_build.State
					returned_build_url = travis_build.Href
					if viper.GetBool("verbose") {
						spew.Dump("travis build", travis_build)
					}
				}
			}

			switch returned_build_status {
			case "received":
				returned_build_status = "running"
			case "created":
				returned_build_status = "running"
			case "started":
				returned_build_status = "running"
			case "passed":
				returned_build_status = "success"
			case "errored":
				returned_build_status = "failed"
			case "failed":
				returned_build_status = "failed"
			default:
				os.Stdout.Sync()
				fmt.Fprintf(os.Stderr, "ERROR: %v \n", "failed to find project with given commit")
				os.Exit(1)
			}

			// url_prefix := fmt.Sprintf("https://travis-ci.org/%s/jobs", travis_build.Repository.Slug)
			// cli_response.JobUrl = strings.Replace(job.Href, "/job", url_prefix, 1)
			url_prefix := fmt.Sprintf("https://travis-ci.org/%s/builds", travis_build.Repository.Slug)
			cli_response.BuildUrl = strings.Replace(returned_build_url, "/build", url_prefix, 1)
			cli_response.BuildStatus = returned_build_status
			return
		}
		cli_proxy_response := retrieveBuildStatus(build)
		// number := cmd.Flag("project")
		// spew.Dump("this is project", number.Value.String())
		// spew.Dump("this is viper project", viper.GetString("project"))
		fmt.Printf(cli_proxy_response.output())
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.PersistentFlags().StringP("project", "p", "", "travis-ci project name")
	statusCmd.PersistentFlags().StringP("commit", "c", "", "travis-ci project commit sha")
	statusCmd.PersistentFlags().StringP("tag", "t", "", "travis-ci project tag")
	statusCmd.PersistentFlags().BoolP("verbose", "v", false, "travis-ci verbose output")
	viper.BindPFlag("project", statusCmd.PersistentFlags().Lookup("project"))
	viper.BindPFlag("commit", statusCmd.PersistentFlags().Lookup("commit"))
	viper.BindPFlag("tag", statusCmd.PersistentFlags().Lookup("tag"))
	viper.BindPFlag("verbose", statusCmd.PersistentFlags().Lookup("verbose"))

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statusCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statusCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
