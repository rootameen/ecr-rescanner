package rescan

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/docker/docker/client"
	"github.com/rootameen/ecr-rescanner/pkg/docker"
	"github.com/rootameen/ecr-rescanner/pkg/ecr"
)

// RescanImage rescans an ECR image if its scan status is SCAN_ELIGIBILITY_EXPIRED or NOT_FOUND.
// If mode is "full-rescan", the image is pulled from ECR, deleted from ECR, and pushed back to ECR.
func RescanImage(image ecr.RepoImage, Uri string, dc *client.Client, authToken string, cfg aws.Config, repo ecr.ECRRepo, mode string, deleteLocal bool) {

	fmt.Println("Pulling image: ", Uri)
	docker.PullImage(dc, authToken, Uri)

	if mode == "full-rescan" {
		fmt.Println("Removing image from ECR: ", Uri)
		ecr.DeleteEcrImage(cfg, repo.RepositoryName, image.ImageDigest, image.ImageTag)

		fmt.Println("Repushing image: ", Uri)
		docker.PushImage(dc, authToken, Uri)

		if deleteLocal {
			fmt.Println("Removing local image: ", Uri)
			docker.RemoveImage(dc, Uri)
		}

	}
}
