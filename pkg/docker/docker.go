package docker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/registry"
	dockerClient "github.com/docker/docker/client"
)

// CreateDockerClient creates a new Docker client using the default environment variables.
func CreateDockerClient() *dockerClient.Client {
	cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal(err)
	}
	return cli
}

// ConfigureECRAuth configures the authentication for an Amazon Elastic Container Registry (ECR) using the specified token.
func ConfigureECRAuth(token string) string {
	ac := registry.AuthConfig{
		Username: "AWS",
		Password: token,
	}

	authData, err := json.Marshal(ac)
	if err != nil {
		log.Fatal(err)
	}

	auth := base64.URLEncoding.EncodeToString(authData)
	return auth
}

// PullImage pulls a Docker image locally from a registry using the specified client, token, and image name.
func PullImage(cli *dockerClient.Client, token string, image string) {

	auth := ConfigureECRAuth(token)
	out, err := cli.ImagePull(context.Background(), image, types.ImagePullOptions{
		RegistryAuth: auth,
	})

	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Image %s pulled successfully\n", image)
	}

	defer out.Close()

	io.Copy(os.Stdout, out)
}

// PushImage pushes a Docker image to a registry using the specified client, token, and image name.
func PushImage(cli *dockerClient.Client, token string, image string) {

	auth := ConfigureECRAuth(token)

	out, err := cli.ImagePush(context.Background(), image, types.ImagePushOptions{
		RegistryAuth: auth,
	})

	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Image %s pushed successfully\n", image)
	}

	defer out.Close()

	io.Copy(os.Stdout, out)
}

// RemoveImage removes a Docker image from the local Docker engine using the specified client and image name.
func RemoveImage(cli *dockerClient.Client, image string) {
	_, err := cli.ImageRemove(context.Background(), image, types.ImageRemoveOptions{})

	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Image %s removed successfully\n", image)
	}
}
