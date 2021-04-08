package record

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

func Chrome() error {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ts := httptest.NewServer(writeHTML(`
		<!doctype html>
		<html class="no-js" lang="">

		<head>
			<meta charset="utf-8">
			<title>Ωmega - Chrome</title>
			<meta name="description" content="Chrome animation recorder">
			<meta name="viewport" content="width=device-width, initial-scale=1">

			<!--<link rel="manifest" href="site.webmanifest"> -->
			<!--<link rel="apple-touch-icon" href="icon.png"> -->
			<!-- Place favicon.ico in the root directory -->

			<!-- <link rel="stylesheet" href="css/normalize.css"> -->
			<!-- <link rel="stylesheet" href="css/style.css"> -->

			<!-- Custom Styles -->
			<style>

			</style>
			<meta name="theme-color" content="#fafafa">
		</head>

		<body>
			<!-- HTML Body -->

			<!-- Custom Script -->
			<script>
				console.log("hello world")
				console.warn("scary warning", 123)
				null.throwsException
			</script>
		</body>

		</html>
	`))
	defer ts.Close()

	exception := make(chan bool, 1)
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *runtime.EventConsoleAPICalled:
			fmt.Printf("* console.%s call:\n", ev.Type)
			for _, arg := range ev.Args {
				fmt.Printf("%s - %s\n", arg.Type, arg.Value)
			}
		case *runtime.EventExceptionThrown:
			// Since ts.URL uses a random port, replace it.
			s := ev.ExceptionDetails.Error()
			s = strings.ReplaceAll(s, ts.URL, "<server>")
			fmt.Printf("* %s \n", s)
			exception <- true
		}
	})

	if err := chromedp.Run(ctx, chromedp.Navigate(ts.URL)); err != nil {
		return err
	}

	<- exception

	return nil
}

func writeHTML(content string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, strings.TrimSpace(content))
	})
}