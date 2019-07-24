package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/apex/log"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gorilla/mux"
)

type YoutubeBackup struct {
	FilePath string
	URL      string
}

var views = template.Must(template.ParseGlob("templates/*.html"))

func main() {
	addr := ":" + os.Getenv("PORT")
	app := mux.NewRouter()
	app.HandleFunc("/", today)
	app.HandleFunc("/v", showVideos)
	if err := http.ListenAndServe(addr, app); err != nil {
		log.WithError(err).Fatal("error listening")
	}
}

func today(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, fmt.Sprintf("/v?date=%s", time.Now().AddDate(0, 0, -1).Format("2006-01-02")), http.StatusFound)
}

func showVideos(w http.ResponseWriter, r *http.Request) {
	date := r.FormValue("date")
	ctx := log.WithFields(log.Fields{
		"date": date,
	})

	tz := "Asia/Singapore"
	_, err := parseDate(date, tz)
	if err != nil {
		ctx.WithError(err).Error("bad date")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cfg, err := external.LoadDefaultAWSConfig(external.WithSharedConfigProfile("mine"))
	if err != nil {
		log.WithError(err).Fatal("failed to load config")
	}
	cfg.Region = "eu-west-1"
	svc := s3.New(cfg)
	req := svc.ListObjectsRequest(&s3.ListObjectsInput{
		Bucket: aws.String("c.prazefarm.co.uk"),
		Prefix: aws.String(date),
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
				log.WithError(err).Fatal("failed to sign url")
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

func parseDate(input string, tz string) (day time.Time, err error) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		log.WithError(err).Error("bad timezone")
		return
	}

	day, err = time.ParseInLocation("2006-01-02", input, loc)
	if err != nil {
		log.WithError(err).Info("bad date")
		return
	}

	return
}
