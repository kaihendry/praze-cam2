package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type YoutubeBackup struct {
	FilePath string
	URL      string
}

var views = template.Must(template.ParseGlob("templates/*.html"))

func main() { log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), routes())) }

func routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", showVideos)
	return mux
}

func showVideos(w http.ResponseWriter, r *http.Request) {
	cfg, err := external.LoadDefaultAWSConfig(external.WithSharedConfigProfile("mine"))
	if err != nil {
		log.Fatalf("failed to load config, %v", err)
	}
	cfg.Region = "eu-west-1"
	svc := s3.New(cfg)
	req := svc.ListObjectsRequest(&s3.ListObjectsInput{
		Bucket: aws.String("c.prazefarm.co.uk"),
		Prefix: aws.String("2019-07-23"),
	})
	p := s3.NewListObjectsPaginator(req)
	var listing []YoutubeBackup
	for p.Next(context.TODO()) {
		page := p.CurrentPage()
		for _, obj := range page.Contents {
			req := svc.GetObjectRequest(&s3.GetObjectInput{
				Bucket: aws.String("c.prazefarm.co.uk"),
				Key:    obj.Key,
			})
			urlStr, err := req.Presign(15 * time.Minute)
			if err != nil {
				log.Fatalf("failed to load make an object request, %v", err)
			}
			listing = append(listing, YoutubeBackup{FilePath: *obj.Key, URL: urlStr})
		}
	}

	if p.Err() != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = views.ExecuteTemplate(w, "index.html", struct {
		Listing []YoutubeBackup
	}{
		listing,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
