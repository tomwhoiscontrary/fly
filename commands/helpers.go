package commands

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/codegangsta/cli"
	"github.com/concourse/atc"
	"github.com/tedsuo/rata"
)

func getBuild(ctx *cli.Context, client *http.Client, reqGenerator *rata.RequestGenerator) atc.Build {
	jobName := ctx.String("job")
	buildName := ctx.String("build")

	if jobName != "" && buildName != "" {
		buildReq, err := reqGenerator.CreateRequest(
			atc.GetJobBuild,
			rata.Params{
				"job_name":   jobName,
				"build_name": buildName,
			},
			nil,
		)
		if err != nil {
			log.Fatalln("failed to create request", err)
		}

		buildResp, err := client.Do(buildReq)
		if err != nil {
			log.Fatalln("failed to get builds:", err)
		}

		if buildResp.StatusCode != http.StatusOK {
			log.Println("bad response when getting build:")
			buildResp.Body.Close()
			buildResp.Write(os.Stderr)
			os.Exit(1)
		}

		var build atc.Build
		err = json.NewDecoder(buildResp.Body).Decode(&build)
		if err != nil {
			log.Fatalln("failed to decode job:", err)
		}

		return build
	} else if jobName != "" {
		jobReq, err := reqGenerator.CreateRequest(
			atc.GetJob,
			rata.Params{"job_name": ctx.String("job")},
			nil,
		)
		if err != nil {
			log.Fatalln("failed to create request", err)
		}

		jobResp, err := client.Do(jobReq)
		if err != nil {
			log.Fatalln("failed to get builds:", err)
		}

		if jobResp.StatusCode != http.StatusOK {
			log.Println("bad response when getting job:")
			jobResp.Body.Close()
			jobResp.Write(os.Stderr)
			os.Exit(1)
		}

		var job atc.Job
		err = json.NewDecoder(jobResp.Body).Decode(&job)
		if err != nil {
			log.Fatalln("failed to decode job:", err)
		}

		if job.NextBuild != nil {
			return *job.NextBuild
		} else if job.FinishedBuild != nil {
			return *job.FinishedBuild
		} else {
			println("job has no builds")
			os.Exit(1)
		}
	} else {
		buildsReq, err := reqGenerator.CreateRequest(
			atc.ListBuilds,
			nil,
			nil,
		)
		if err != nil {
			log.Fatalln("failed to create request", err)
		}

		buildsResp, err := client.Do(buildsReq)
		if err != nil {
			log.Fatalln("failed to get builds:", err)
		}

		if buildsResp.StatusCode != http.StatusOK {
			log.Println("bad response when getting builds:")
			buildsResp.Body.Close()
			buildsResp.Write(os.Stderr)
			os.Exit(1)
		}

		var builds []atc.Build
		err = json.NewDecoder(buildsResp.Body).Decode(&builds)
		if err != nil {
			log.Fatalln("failed to decode builds:", err)
		}

		for _, build := range builds {
			if build.JobName == "" {
				return build
			}
		}

		println("no builds")
		os.Exit(1)
	}

	panic("unreachable")
}
