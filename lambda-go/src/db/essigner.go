/**
 * https://blog.linkode.co.jp/entry/2020/04/22/093502
 */
package db

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	signer "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/hirosato/wcs/env"
)

type amazonESTransport struct {
	awsSigner *signer.Signer
}

func NewAmazonESTransport() *amazonESTransport {
	t := new(amazonESTransport)
	// 環境変数から認証情報を取得し、AWS署名バージョン4の署名者を作成
	t.awsSigner = signer.NewSigner(credentials.NewEnvCredentials())
	return t
}

func (t *amazonESTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if env.IsLocal {
		return http.DefaultTransport.RoundTrip(req)
	}

	const service = "es"

	if h, ok := req.Header["Authorization"]; ok && len(h) > 0 && strings.HasPrefix(h[0], "AWS4") {
		log.Println("リクエストは署名済みなので署名処理をスキップします.")
		return http.DefaultTransport.RoundTrip(req)
	}

	req.URL.Scheme = "https"

	now := time.Now()
	req.Header.Set("Date", now.Format(time.RFC3339))
	log.Printf("署名日時: %+v", req)

	var err error
	switch req.Body {
	case nil:
		log.Println("Bodyなしの署名を行います.")
		_, err = t.awsSigner.Sign(req, nil, service, env.Region, now)
	default:
		var b []byte
		b, err = ioutil.ReadAll(req.Body)
		if err == nil {
			req.Body = ioutil.NopCloser(bytes.NewReader(b))
			log.Println("Body付きの署名を行います.")
			_, err = t.awsSigner.Sign(req, bytes.NewReader(b), service, env.Region, now)
		}
	}
	if err != nil {
		log.Printf("署名中にエラーが発生しました: '%s'", err)
		return nil, err
	}

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		log.Printf("http.DefaultTransport.RoundTrip() に失敗しました.\n\n\tResponse: %+v\n\n\tError: '%s'", resp, err)
		return resp, err
	}

	return resp, nil
}
