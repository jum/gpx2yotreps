package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"
)

var (
	outDir = flag.String("outdir", "msgs", "directory to store email messages")
)

type gpx struct {
	XMLName xml.Name `xml:"gpx"`
	Rte     struct {
		Name  string `xml:"name"`
		Rtept []struct {
			Lat  float64   `xml:"lat,attr"`
			Lon  float64   `xml:"lon,attr"`
			Name string    `xml:"name"`
			Cmt  string    `xml:"cmt"`
			Time time.Time `xml:"time"`
		} `xml:"rtept"`
	} `xml:"rte"`
}

var mailMsg = template.Must(template.New("mail").Funcs(template.FuncMap{
	"subjTime": func(value time.Time) string {
		return value.Format("2006/01/02 15:04:05")
	},
	"timeTime": func(value time.Time) string {
		return value.Format("2006/01/02 15:04")
	},
	"RFC850Time": func(value time.Time) string {
		return value.Format(time.RFC850)
	},
	"fmtLat": func(value float64) string {
		s := 'N'
		if value < 0 {
			s = 'S'
			value *= -1
		}
		deg := int(value)
		minutes := (value - float64(deg)) * 60
		return fmt.Sprintf("%d-%.2f%c", deg, minutes, s)
	},
	"fmtLon": func(value float64) string {
		s := 'E'
		if value < 0 {
			s = 'W'
			value *= -1
		}
		deg := int(value)
		minutes := (value - float64(deg)) * 60
		return fmt.Sprintf("%0d-%.2f%c", deg, minutes, s)
	},
}).Parse(`To: jum@anubis.han.de
Subject: YotReps: {{subjTime .Time}}
Date: {{RFC850Time .Time}}

AIRMAIL YOTREPS
IDENT: DJOE
TIME: {{timeTime .Time}}
LATITUDE: {{fmtLat .Lat}}
LONGITUDE: {{fmtLon .Lon}}
COMMENT: {{.Cmt}}
`))

func main() {
	flag.Parse()
	err := os.MkdirAll(*outDir, 0755)
	if err != nil {
		panic(err)
	}
	li, err := os.Create(filepath.Join(*outDir, "MsgList.txt"))
	if err != nil {
		panic(err)
	}
	msgNum := 0
	for _, fname := range flag.Args() {
		//fmt.Printf("File %s:\n", fname)
		f, err := os.Open(fname)
		if err != nil {
			panic(err)
		}
		dec := xml.NewDecoder(f)
		var g gpx
		err = dec.Decode(&g)
		if err != nil {
			panic(err)
		}
		//fmt.Printf("g.Rte.Rtept %#v\n", g.Rte.Rtept)
		for _, p := range g.Rte.Rtept {
			msgName := fmt.Sprintf("%d_DJOE.msg", msgNum)
			msgNum++
			msgFile, err := os.Create(filepath.Join(*outDir, msgName))
			if err != nil {
				panic(err)
			}
			err = mailMsg.Execute(msgFile, &p)
			if err != nil {
				panic(err)
			}
			err = msgFile.Close()
			if err != nil {
				panic(err)
			}
			fmt.Fprintf(li, "a;Posted;%s;rest\n", msgName[0:len(msgName)-4])
		}
		f.Close()
	}
	err = li.Close()
	if err != nil {
		panic(err)
	}
}
