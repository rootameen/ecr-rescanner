package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/rootameen/ecr-rescanner/pkg/docker"
	"github.com/rootameen/ecr-rescanner/pkg/ecr"
	"github.com/rootameen/ecr-rescanner/pkg/rescan"
)

func main() {

	// flags

	ecrProfile := flag.String("ecrProfile", "", "AWS Profile to use which contains ECR Repos")
	ecrImageRegistry := flag.String("ecrImageRegistry", "", "ECR Image Registry to scan, e.g. 424851304182")
	mode := flag.String("mode", "pull-only", "mode: pull-only, full-rescan")
	deleteLocal := flag.Bool("deleteLocal", false, "delete local images after pushing to ECR (default false)")

	flag.Parse()

	var cfg aws.Config
	var err error

	// ECR AWS Account Auth
	if *ecrProfile == "" {
		cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion("eu-central-1"))
		if err != nil {
			log.Fatalf("Unable to load SDK config, %v", err)
		}

	} else {
		cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion("eu-central-1"), config.WithSharedConfigProfile(*ecrProfile))
		if err != nil {
			log.Fatalf("Unable to load SDK config, %v", err)
		}
	}

	// authenticate to ECR
	endpoint, authToken := ecr.AuthenticateToEcr(cfg, *ecrImageRegistry)
	endpointUri := strings.Replace(endpoint, "https://", "", 1)

	// create docker client
	dc := docker.CreateDockerClient()

	// generate list of all images in ECR
	fmt.Println("Generating list of all images in ECR...")
	ecrRepos := ecr.GenerateEcrImageList(cfg)

	counter := 0
	for _, repo := range ecrRepos {
		for _, image := range repo.RepoImages {
			if image.ScanStatus == "SCAN_ELIGIBILITY_EXPIRED" || image.ScanStatus == "NOT_FOUND" {
				counter++
				switch image.ImageTag {
				case "untagged":
					Uri := endpointUri + "/" + repo.RepositoryName + "@" + image.ImageDigest
					// for untagged images, only pull the image locally without repushing it for rescan
					docker.PullImage(dc, authToken, Uri)
				default: // tagged images
					Uri := endpointUri + "/" + repo.RepositoryName + ":" + image.ImageTag
					rescan.RescanImage(image, Uri, dc, authToken, cfg, repo, *mode, *deleteLocal)
				}
			}
		}
	}

	log.Printf("Total images without active inspector scans: %d", counter)

}
