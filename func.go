package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	fdk "github.com/fnproject/fdk-go"
	"github.com/jung-kurt/gofpdf"
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/objectstorage"
)

func main() {
	fdk.Handle(fdk.HandlerFunc(text2PDF))
}

const privateKeyFolder string = "/function"

func text2PDF(ctx context.Context, in io.Reader, out io.Writer) {

	//log.Println("Function handler invoked ", time.Now())

	fnCtx := fdk.GetContext(ctx)

	tenancy := fnCtx.Config()["TENANT_OCID"]
	user := fnCtx.Config()["USER_OCID"]
	region := fnCtx.Config()["REGION"]
	fingerprint := fnCtx.Config()["FINGERPRINT"]
	privateKeyName := fnCtx.Config()["PRIVATE_KEY_NAME"]
	privateKeyLocation := privateKeyFolder + "/" + privateKeyName
	passphrase := fnCtx.Config()["PASSPHRASE"]
	namespace := fnCtx.Config()["NAMESPACE"]
	bucketName := fnCtx.Config()["BUCKET_NAME"]

	log.Println("TENANT_OCID ", tenancy)
	log.Println("USER_OCID ", user)
	log.Println("REGION ", region)
	log.Println("FINGERPRINT ", fingerprint)
	log.Println("PRIVATE_KEY_NAME ", privateKeyName)
	log.Println("PRIVATE_KEY_LOCATION ", privateKeyLocation)
	log.Println("NAMESPACE ", namespace)
	log.Println("BUCKET_NAME ", bucketName)

	buf := new(bytes.Buffer)
	buf.ReadFrom(in)
	textFileName := buf.String()
	log.Println("Text file name", textFileName)

	if textFileName == "" {
		resp := FailedResponse{Message: "File name empty. Pass name of a valid .txt file in bucket - " + bucketName, Error: ""}
		log.Println("File name empty")
		json.NewEncoder(out).Encode(resp)
		return
	}

	nameWithoutType := strings.Split(textFileName, ".")[0]
	//log.Println("File name without type", nameWithoutType)
	opFileName := nameWithoutType + ".pdf"
	tmpFileLocation := "/tmp/" + opFileName

	privateKey, err := ioutil.ReadFile(privateKeyLocation)
	if err == nil {
		log.Println("read private key from ", privateKeyLocation)
	} else {
		resp := FailedResponse{Message: "Unable to read private Key", Error: err.Error()}
		log.Println(resp.toString())
		json.NewEncoder(out).Encode(resp)
		return
	}

	rawConfigProvider := common.NewRawConfigurationProvider(tenancy, user, region, fingerprint, string(privateKey), common.String(passphrase))
	osclient, err := objectstorage.NewObjectStorageClientWithConfigurationProvider(rawConfigProvider)

	if err != nil {
		resp := FailedResponse{Message: "Problem getting Object Store Client handle", Error: err.Error()}
		log.Println(resp.toString())
		json.NewEncoder(out).Encode(resp)
		return
	}

	log.Println("Reading text file " + textFileName + " from storage bucket " + bucketName)

	req := objectstorage.GetObjectRequest{NamespaceName: common.String(namespace), BucketName: common.String(bucketName), ObjectName: common.String(textFileName)}
	resp, err := osclient.GetObject(context.Background(), req)

	if err != nil {
		resp := FailedResponse{Message: "Could not read file " + textFileName + " from bucket " + bucketName, Error: err.Error()}
		log.Println(resp.toString())
		json.NewEncoder(out).Encode(resp)
		return
	}

	buf = new(bytes.Buffer)
	buf.ReadFrom(resp.Content)

	text := buf.String()

	//log.Println("Got text from file", text)

	err = textToPDF(text, tmpFileLocation)
	if err == nil {
		log.Println("PDF Written to " + tmpFileLocation)
	} else {
		resp := FailedResponse{Message: "Failed to write PDF", Error: err.Error()}
		log.Println(resp.toString())
		json.NewEncoder(out).Encode(resp)
		return
	}

	defer func() {
		fileErr := os.Remove(tmpFileLocation)
		if fileErr == nil {
			log.Println("Deleted temp file", tmpFileLocation)
		} else {
			log.Println("Error removing temp file", fileErr.Error())
		}
	}()

	file, err := os.Open(tmpFileLocation)
	if err != nil {

		resp := FailedResponse{Message: "failed to read PDF from " + tmpFileLocation, Error: err.Error()}
		log.Println(resp.toString())
		json.NewEncoder(out).Encode(resp)
		return
	}
	defer file.Close()

	info, err := file.Stat()
	//log.Println("PDF File size -- ", info.Size())

	putReq := objectstorage.PutObjectRequest{ContentLength: common.Int64(info.Size()), PutObjectBody: file, NamespaceName: common.String(namespace), BucketName: common.String(bucketName), ObjectName: common.String(opFileName)}
	_, err = osclient.PutObject(context.Background(), putReq)

	if err == nil {
		msg := "PDF " + opFileName + " written to storage bucket - " + bucketName
		log.Println(msg)
		out.Write([]byte(msg))
	} else {
		resp := FailedResponse{Message: "Failed to write PDF to bucket", Error: err.Error()}
		log.Println(resp.toString())
		json.NewEncoder(out).Encode(resp)
		return
	}
}

func textToPDF(text, tmpFileLocation string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Times", "", 12)
	pdf.MultiCell(0, 5, string(text), "", "", false)
	return pdf.OutputFileAndClose(tmpFileLocation)
}

//FailedResponse ...
type FailedResponse struct {
	Message string
	Error   string
}

func (response FailedResponse) toString() string {
	return response.Message + " due to " + response.Error
}
