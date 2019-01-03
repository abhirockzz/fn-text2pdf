# Function for converting text to PDF

This Function converts a text file to PDF. It reads a text (`.txt`) file from a Oracle Cloud Infrastructure Object Storage Bucket, converts it into PDF and stores the converted file in the same bucket (with a `.pdf` extension)

- It's written in Go and uses [gofpdf](https://github.com/jung-kurt/gofpdf) for text to PDF conversion 
- Uses the [OCI Go SDK](https://github.com/oracle/oci-go-sdk) to execute Object Storage read and write operations
- A custom Dockerfile is used to build the function

## Pre-requisites

- Start by cloning this repository
- [Create Oracle Cloud Infrastructure Object Storage bucket](https://docs.cloud.oracle.com/iaas/Content/Object/Tasks/managingbuckets.htm#usingconsole)
- Collect the following information for you OCI tenancy (you'll need these in subsequent steps) - Tenancy OCID, User OCID of a user in the tenancy, OCI private key, OCI public key passphrase, OCI region, Object Storage namespace and name of the bucket you just created
- Copy your OCI private key to folder. If you don't already have one, [please follow the documentation](https://docs.cloud.oracle.com/iaas/Content/API/Concepts/apisigningkey.htm#How)


### Switch to correct context

- `fn use context <your context name>`
- Check using `fn ls apps`

## Create application

`fn create app text2pdf --config TENANT_OCID=<TENANT_OCID> --config USER_OCID=<USER_OCID> --config FINGERPRINT=<FINGERPRINT> --config PASSPHRASE=<PASSPHRASE> --config REGION=<REGION> --config PRIVATE_KEY_NAME=<PRIVATE_KEY_NAME> --config NAMESPACE=<NAMESPACE> --config BUCKET_NAME=<BUCKET_NAME>`

e.g.

`fn create app text2pdf --config TENANT_OCID=ocid1.tenancy.oc1..aaaaaaaaydrjm77otncda2xn7qtv7l3hqnd3zxn2u6siwdhniibwfv4wwhta --config USER_OCID=ocid1.user.oc1..aaaaaaaa4seqx6jeyma46ldy4cbuv35q4l26scz5p4rkz3rauuoioo26qwmq --config FINGERPRINT=41:82:5f:44:ca:a1:2e:58:d2:63:6a:af:52:d5:3d:04 --config PASSPHRASE=1987 --config REGION=us-phoenix-1 --config PRIVATE_KEY_NAME=oci_private_key.pem --config NAMESPACE=oracle-functions --config BUCKET_NAME=test-bucket`

### Check

`fn inspect app text2pdf`

## Deploy the application

- `cd fn-text2pdf` 
- `fn -v deploy --app text2pdf --build-arg PRIVATE_KEY_NAME=<private_key_name>` e.g. `fn -v deploy --app text2pdf --build-arg PRIVATE_KEY_NAME=oci_private_key.pem`

## Test

A sample text file (`lorem.txt`) has been provided to test the function

- Upload file to your object storage bucket,
- and then test using `echo -n 'lorem.txt' | fn invoke text2pdf convert`

If successful, you should see the following output - `PDF lorem.pdf written to storage bucket - test-bucket`. Now, you can download the PDF (`lorem.pdf`) from your Object Storage bucket
