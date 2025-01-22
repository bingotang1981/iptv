package liveurls

import (
	// "context"
	"fmt"
	"io"
	// "net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Tptv struct{}

func (i *Tptv) HandleMainRequest(w http.ResponseWriter, r *http.Request, cdn string, id string, playseek string) {

	//Compatible with TVBox
	cdn = strings.ReplaceAll(cdn, "/PLTV/8888/", "/PLTV/")

	//If playback, replace PLTV with TVOD
	if playseek != "" {
		cdn = cdn + "?playseek=" + playseek
		cdn = strings.ReplaceAll(cdn, "/PLTV/", "/TVOD/")
	}

	data, redirectURL, err := getTptvHTTPResponse(cdn)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	redirectPrefix := redirectURL[:strings.LastIndex(redirectURL, "/")+1]

	// 替换TS文件的链接
	golang := "http://" + r.Host + r.URL.Path

	re := regexp.MustCompile(`((?i).*?\.ts)`)
	data = re.ReplaceAllStringFunc(data, func(match string) string {
		return golang + "?ts=" + redirectPrefix + match
	})

	// 将&替换为$
	data = strings.ReplaceAll(data, "&", MYSEPERETOR)

	w.Header().Set("Content-Disposition", "attachment;filename="+id)
	w.WriteHeader(http.StatusOK) // Set the status code to 200
	w.Write([]byte(data))        // Write the response body

	// //Build starturl from cdn and id
	// myid := strings.ReplaceAll(id, ".m3u8", "")
	// startUrl := "http://gslbserv.itv.cmvideo.cn:80/" + myid + "/1.m3u8?channel-id=" + cdn + "&Contentid=" + myid + "&livemode=1&stbId=003803ff00010060180758b42d777238"

	// data, redirectURL, err := getHTTPResponse(startUrl)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	// redirectPrefix := redirectURL[:strings.LastIndex(redirectURL, "/")+1]

	// // 替换TS文件的链接
	// golang := "http://" + r.Host + r.URL.Path

	// re := regexp.MustCompile(`((?i).*?\.ts)`)
	// data = re.ReplaceAllStringFunc(data, func(match string) string {
	// 	return golang + "?ts=" + redirectPrefix + match
	// })

	// // 将&替换为$
	// data = strings.ReplaceAll(data, "&", MYSEPERETOR)

	// w.Header().Set("Content-Disposition", "attachment;filename="+id)
	// w.WriteHeader(http.StatusOK) // Set the status code to 200
	// w.Write([]byte(data))        // Write the response body
}

func (i *Tptv) HandleTsRequest(w http.ResponseWriter, ts string) {

	// 将$替换回&
	ts = strings.ReplaceAll(ts, MYSEPERETOR, "&")

	// Read one piece and then write one piece
	w.Header().Set("Content-Type", "video/MP2T")
	_, _, err := handleTptvTsHTTPResponse(ts, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getTptvHTTPResponse(requestURL string) (string, string, error) {

	redirectURL := ""

	// 创建一个新的 HTTP 客户端并设置连接超时为5秒
	client := &http.Client{
		Timeout: 10 * time.Second,

		//Get redirect url via automatic redirect
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			redirectURL = req.URL.String()
			fmt.Println("RedirectURL: " + redirectURL)
			return nil
		},
	}

	fmt.Println("RequestURL: " + requestURL)

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "okhttp")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", http.ErrServerClosed
	}

	body, _ := readResponseBodyTptv(resp)

	if redirectURL == "" {
		redirectURL = requestURL
	}

	return body, redirectURL, nil
}

func handleTptvTsHTTPResponse(requestURL string, w http.ResponseWriter) (string, string, error) {

	// 创建一个新的 HTTP 客户端并设置连接超时为5秒
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	fmt.Println("RequestURL: " + requestURL)

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "okhttp")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", http.ErrServerClosed
	}

	w.WriteHeader(http.StatusOK)    // Set the status code to 200
	buffer := make([]byte, 40*1024) // 40KB buffer
	for {
		n, myerr := resp.Body.Read(buffer)
		if myerr == io.EOF {
			break // End of file reached
		}
		if myerr != nil {
			fmt.Println("Error reading response body:", myerr)
			err = myerr
			break
		}
		if n > 0 {
			w.Write([]byte(buffer[:n])) // Write the response body
		}
	}

	if err != nil {
		
		return "", "", err
	}

	return "", "", nil
}

func readResponseBodyTptv(resp *http.Response) (string, error) {
	var builder strings.Builder
	_, err := io.Copy(&builder, resp.Body)
	if err != nil {
		return "", err
	}
	return builder.String(), nil
}
