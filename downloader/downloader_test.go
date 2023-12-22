package downloader

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestGetInfosFromGithubLink(t *testing.T) {
	tests := []struct {
		url           string
		expectedErr   error
		expectedOwner string
		expectedRepo  string
	}{
		{
			url:           "https://github.com/owner/repo",
			expectedErr:   nil,
			expectedOwner: "owner",
			expectedRepo:  "repo",
		},
		{
			url:           "https://github.com/owner/repo/subpath",
			expectedErr:   nil,
			expectedOwner: "owner",
			expectedRepo:  "repo",
		},
		{
			url:           "https://github.com/owner",
			expectedErr:   fmt.Errorf("Could not find owner and repo in https://github.com/owner"),
			expectedOwner: "",
			expectedRepo:  "",
		},
		{
			url:           "https://example.com",
			expectedErr:   fmt.Errorf("Could not find owner and repo in https://example.com"),
			expectedOwner: "",
			expectedRepo:  "",
		},
	}

	for _, test := range tests {
		owner, repo, err := getInfosFromGithubLink(test.url)
		if err != nil && test.expectedErr == nil {
			t.Errorf("Unexpected error for URL %s: %v", test.url, err)
		} else if err == nil && test.expectedErr != nil {
			t.Errorf("Expected error for URL %s: %v", test.url, test.expectedErr)
		} else if err != nil && test.expectedErr != nil && err.Error() != test.expectedErr.Error() {
			t.Errorf("Unexpected error message for URL %s: got %v, want %v", test.url, err, test.expectedErr)
		}

		if owner != test.expectedOwner {
			t.Errorf("Unexpected owner for URL %s: got %s, want %s", test.url, owner, test.expectedOwner)
		}

		if repo != test.expectedRepo {
			t.Errorf("Unexpected repo for URL %s: got %s, want %s", test.url, repo, test.expectedRepo)
		}
	}
}

func Test_getGithubReleaseAssetUrl(t *testing.T) {
	type args struct {
		owner   string
		repo    string
		filter  string
		httpGet func(url string) (*http.Response, error)
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "",
			args: args{
				owner:  "FiloSottile",
				repo:   "age",
				filter: "age-v(.*)-windows-amd64.zip",
				httpGet: func(url string) (*http.Response, error) {
					if url != "https://api.github.com/repos/FiloSottile/age/releases/latest" {
						t.Errorf("Unexpected URL: %s", url)
					}
					body := `{
						"draft": false,
						"prerelease": false,
						"assets": [
						  {
							"name": "age-v1.1.1-darwin-amd64.tar.gz",
							"browser_download_url": "https://github.com/FiloSottile/age/releases/download/v1.1.1/age-v1.1.1-darwin-amd64.tar.gz"
						  },
						  {
							"name": "age-v1.1.1-darwin-arm64.tar.gz",
							"browser_download_url": "https://github.com/FiloSottile/age/releases/download/v1.1.1/age-v1.1.1-darwin-arm64.tar.gz"
						  },
						  {
							"name": "age-v1.1.1-windows-amd64.zip",
							"browser_download_url": "https://github.com/FiloSottile/age/releases/download/v1.1.1/age-v1.1.1-windows-amd64.zip"
						  }
						],
						"tarball_url": "https://api.github.com/repos/FiloSottile/age/tarball/v1.1.1",
						"zipball_url": "https://api.github.com/repos/FiloSottile/age/zipball/v1.1.1"
					  }`
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(strings.NewReader(body)),
					}, nil
				},
			},
			want: "https://github.com/FiloSottile/age/releases/download/v1.1.1/age-v1.1.1-windows-amd64.zip",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getHttp = tt.args.httpGet
			got, err := getGithubReleaseAssetUrl(tt.args.owner, tt.args.repo, tt.args.filter)
			if (err != nil) != tt.wantErr {
				t.Errorf("getGithubReleaseAssetUrl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getGithubReleaseAssetUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}
