package ecr

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/remeh/sizedwaitgroup"
)

// ECRRepo represents an ECR repository
type ECRRepo struct {
	RepositoryName string
	RepositoryArn  string
	RepoImages     []RepoImage
}

// RepoImage represents an image in an ECR repository
type RepoImage struct {
	ImageDigest string
	ImageTag    string
	ScanStatus  string
}

// AuthenticateToEcr authenticates to ECR and returns the endpoint and auth token
func AuthenticateToEcr(cfg aws.Config, reg string) (endpoint string, authToken string) {
	client := ecr.NewFromConfig(cfg)
	auth, err := client.GetAuthorizationToken(context.TODO(), &ecr.GetAuthorizationTokenInput{
		RegistryIds: []string{reg},
	})
	if err != nil {
		fmt.Println(err)
	}

	endpoint = *auth.AuthorizationData[0].ProxyEndpoint
	authToken = *auth.AuthorizationData[0].AuthorizationToken

	// Decode the auth token
	dec, err := base64.URLEncoding.DecodeString(authToken)
	if err != nil {
		fmt.Printf("Failed to decode ECR Token: %s", err)
	}

	// Split the token into endpoint and token parts
	token := strings.Split(string(dec), ":")
	if len(token) != 2 {
		fmt.Printf("Unexpected ECR Token format")
	}

	return endpoint, token[1]
}

// GenerateEcrImageList generates a list of ECR repositories and their images
func GenerateEcrImageList(cfg aws.Config) []ECRRepo {
	client := ecr.NewFromConfig(cfg)

	var ecrRepo []ECRRepo
	var scanStatus string

	// Get a list of ECR repositories
	repos, err := client.DescribeRepositories(context.TODO(), &ecr.DescribeRepositoriesInput{})
	if err != nil {
		fmt.Printf("Unable to list repositories, %v", err.Error())
	}

	// For each repository, get a list of images
	for _, repo := range repos.Repositories {

		repoArn := *repo.RepositoryArn

		if err != nil {
			fmt.Printf("Unable to list repositories, %v", err.Error())
		}

		imgs, _ := client.ListImages(context.TODO(), &ecr.ListImagesInput{
			RepositoryName: aws.String(*repo.RepositoryName),
		})

		var repoImages []RepoImage
		swg := sizedwaitgroup.New(5)

		// For each image, get its scan status and add it to the list of images for the repository
		for _, image := range imgs.ImageIds {
			swg.Add()

			go func(image types.ImageIdentifier) {
				defer swg.Done()
				if image.ImageTag == nil {
					image.ImageTag = aws.String("untagged")
				}
				if image.ImageDigest == nil {
					image.ImageDigest = aws.String("unknown")
				}

				// Get the scan status for the image
				scanFindingsInput := &ecr.DescribeImageScanFindingsInput{
					RepositoryName: aws.String(*repo.RepositoryName),
					ImageId: &types.ImageIdentifier{
						ImageTag:    aws.String(*image.ImageTag),
						ImageDigest: aws.String(*image.ImageDigest),
					},
				}

				scanFindingsOutput, _ := client.DescribeImageScanFindings(context.TODO(), scanFindingsInput)

				if scanFindingsOutput == nil {
					scanStatus = "NOT_FOUND"
				} else {
					scanStatus = string(scanFindingsOutput.ImageScanStatus.Status)
				}

				repoImages = append(repoImages, RepoImage{*image.ImageDigest, *image.ImageTag, scanStatus})
			}(image)
		}

		swg.Wait()
		ecrRepo = append(ecrRepo, ECRRepo{*repo.RepositoryName, repoArn, repoImages})
	}
	return ecrRepo
}

// DeleteEcrImage deletes an image from an ECR repository
func DeleteEcrImage(cfg aws.Config, repoName string, imageDigest string, imageTag string) {
	client := ecr.NewFromConfig(cfg)

	// Delete the image from the repository
	_, err := client.BatchDeleteImage(context.TODO(), &ecr.BatchDeleteImageInput{
		RepositoryName: aws.String(repoName),
		ImageIds: []types.ImageIdentifier{
			{
				ImageDigest: aws.String(imageDigest),
				ImageTag:    aws.String(imageTag),
			},
		},
	})
	if err != nil {
		fmt.Println(err)
	}
}
