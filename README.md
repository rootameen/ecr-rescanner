# ECR Rescanner

ECR Rescanner targets users of AWS Elastic Container Registry (ECR) and AWS Inspector. Inspector currently has a limitation where [only images uploaded up to 30 days before its activation to scan ECR images are scanned](https://docs.aws.amazon.com/inspector/latest/user/scanning-ecr.html#:~:text=When%20you%20first%20activate%20Amazon%20ECR%20scanning%2C%20Amazon%20Inspector%20scans%20eligible%20images%20pushed%20in%20the%20last%2030%20days). ECR Rescanner reuploads the images which have expired or non-existing scan results, allowing for a more comprehensive security scan of your ECR repositories.

## How it works

As of now, the only way to trigger a rescan of an image on ECR is to delete it and re-upload it again. This tool does that, by pulling the detected target images - ones that currently do not have scan results - to the local system, remove them from ECR, then re-upload them again to initiate their scan.

## Installation

To install ECR Rescanner, you will need to have Go installed on your system. Once you have Go installed, you can run the following command to install ECR Rescanner:

```
go get github.com/rootameen/ecr-rescanner
```

## Usage

To use ECR Rescanner, you will need to have AWS credentials set up on your system. Once you have your credentials set up, you can run the following command to scan all images in your ECR repository:

```
ecr-rescanner -ecrProfile <profile-name> -ecrImageRegistry <registry-id> -mode <mode> -deleteLocal <true/false>
```

The `-ecrProfile` flag specifies the name of the AWS profile to use for authentication. `-ecrImageRegistry` sets the ECR Registry ID. The`-mode` flag specifies the mode of operation for the tool. The available modes are:

* pull-only (default): only pulls the eligible images for rescanning to the local system
* full-rescan: This mode pulls all eligible images in the repository, deletes them from ECR, then reuploads them again to initiate their scan

The `-deleteLocal` (bool) flag specifies whether to delete the local copy of the image after it has been uploaded to ECR.
