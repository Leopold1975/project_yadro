package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/Leopold1975/yadro_app/internal/app"
	"github.com/Leopold1975/yadro_app/internal/controller/httpserver"
	"github.com/Leopold1975/yadro_app/internal/pkg/config"
	"github.com/stretchr/testify/suite"
)

var containerManager string

var (
	imageName     = "comics_db_test"
	containerName = "comics_db_test_c"
	configFile    = "config_test.yaml"
)

func TestMain(m *testing.M) {
	containerManager = os.Getenv("CM")

	if containerManager == "" {
		os.Exit(0)
	}

	if containerManager != "podman" && containerManager != "docker" {
		os.Exit(1)
	}

	os.Exit(m.Run())
}

type ComicsSuite struct {
	suite.Suite
	cancel context.CancelFunc
	cfg    config.Config
}

func (cs *ComicsSuite) SetupSuite() {
	buildCmd := exec.Command(containerManager, "build", "-t", imageName, "-f", "./Dockerfile.test", ".")
	buildCmd.Stderr = os.Stderr
	buildCmd.Stdout = os.Stdout

	if err := buildCmd.Run(); err != nil {
		cs.T().Fatalf("build image error %v", err)
	}

	if containerManager == "podman" {
		imageName += ":latest"
	}

	runCmd := exec.Command(containerManager, "run", "--name", containerName, "-d", "-p", "7788:5432", imageName)
	runCmd.Stderr = os.Stderr
	runCmd.Stdout = os.Stdout

	if err := runCmd.Run(); err != nil {
		cs.T().Fatalf("run container error %v", err)
	}

	time.Sleep(time.Second) // ждём запуска контейнеров

	cfg, err := config.New(configFile)
	cs.Require().NoError(err)

	cs.cfg = cfg

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)

	cs.cancel = cancel

	go app.Run(ctx, cfg, false)

	time.Sleep(time.Second * 2) // ждем запуска сервиса
}

func (cs *ComicsSuite) TearDownSuite() {
	cs.cancel()

	if containerManager != "" {
		stopCmd := exec.Command("make", "clean")
		stopCmd.Stderr = os.Stderr
		stopCmd.Stdout = os.Stdout
		if err := stopCmd.Run(); err != nil {
			cs.T().Fatalf("stop container error %v", err)
		}
	}
}

func (cs *ComicsSuite) TestBasic() {
	req := httpserver.LoginRequest{
		Username: "admin",
		Password: "admin",
	}

	jsonData, err := json.Marshal(req)
	cs.Require().NoError(err)

	cs.T().Log("Running basic test...")
	cs.T().Log("Sending POST request to login endpoint...")

	resp, err := http.Post("http://"+cs.cfg.Server.Addr+"/login", "application/json", bytes.NewBuffer(jsonData))
	cs.Require().NoError(err)

	defer resp.Body.Close()

	cs.Require().Equal(http.StatusOK, resp.StatusCode)

	h := resp.Header.Get("Authorization")

	cs.Require().True(len(h) >= 1)
	cs.Require().True(strings.HasPrefix(h, "Bearer "))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	reqHttp, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://"+cs.cfg.Server.Addr+"/update", nil)

	cs.Require().NoError(err)

	reqHttp.Header.Set("Authorization", h)
	reqHttp.Header.Set("Content-Type", "application/json")

	cs.T().Log("Sending POST request to update endpoint...")
	resp, err = http.DefaultClient.Do(reqHttp)
	if errors.Is(err, context.DeadlineExceeded) {
		for i := 0; i < 3; i++ {
			cs.T().Log("Retrying...")

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
			defer cancel()

			reqHttp, err = http.NewRequestWithContext(ctx, http.MethodPost, "http://"+cs.cfg.Server.Addr+"/update", nil)
			cs.Require().NoError(err)

			reqHttp.Header.Set("Authorization", h)
			reqHttp.Header.Set("Content-Type", "application/json")

			resp, err = http.DefaultClient.Do(reqHttp)
			if !errors.Is(err, context.DeadlineExceeded) {
				break
			}
		}
	}
	cs.Require().NoError(err)

	defer resp.Body.Close()

	result := struct {
		New   int `json:"new"`
		Total int `json:"total"`
	}{}

	dec := json.NewDecoder(resp.Body)

	err = dec.Decode(&result)
	cs.Require().NoError(err)

	cs.Require().Greater(result.Total, 2000)

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	qe := url.QueryEscape("apple doctor")

	reqHttp, err = http.NewRequestWithContext(ctx, http.MethodGet, "http://"+cs.cfg.Server.Addr+"/pics?search="+qe, nil)

	cs.Require().NoError(err)

	reqHttp.Header.Set("Authorization", h)
	reqHttp.Header.Set("Content-Type", "application/json")

	cs.T().Log("Sending GET request to pics endpoint...")
	resp, err = http.DefaultClient.Do(reqHttp)

	cs.Require().NoError(err)
	defer resp.Body.Close()

	resultPics := struct {
		URLs []string `json:"urls"`
	}{}

	dec = json.NewDecoder(resp.Body)

	err = dec.Decode(&resultPics)
	cs.Require().NoError(err)

	cs.Require().Equal(10, len(resultPics.URLs))

	for i, u := range resultPics.URLs {
		if u == "https://imgs.xkcd.com/comics/an_apple_a_day.png" {
			break
		}
		if i == 9 {
			cs.T().Error("expected comics not found")
		}
	}
}

func (cs *ComicsSuite) TestUserScenarios() {
	req := httpserver.LoginRequest{
		Username: "user1",
		Password: "123456",
	}

	jsonData, err := json.Marshal(req)
	cs.Require().NoError(err)

	cs.T().Log("Running user test...")
	cs.T().Log("Sending POST request to login endpoint...")

	resp, err := http.Post("http://"+cs.cfg.Server.Addr+"/login", "application/json", bytes.NewBuffer(jsonData))
	cs.Require().NoError(err)

	defer resp.Body.Close()

	cs.Require().Equal(http.StatusOK, resp.StatusCode)

	h := resp.Header.Get("Authorization")

	cs.Require().True(len(h) >= 1)
	cs.Require().True(strings.HasPrefix(h, "Bearer "))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	reqHttp, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://"+cs.cfg.Server.Addr+"/update", nil)

	cs.Require().NoError(err)

	reqHttp.Header.Set("Authorization", h)
	reqHttp.Header.Set("Content-Type", "application/json")

	cs.T().Log("Sending POST request to update endpoint...")
	resp, err = http.DefaultClient.Do(reqHttp)

	cs.Require().Equal(http.StatusForbidden, resp.StatusCode)
	cs.Require().NoError(err)
	defer resp.Body.Close()

	qe := url.QueryEscape("apple doctor")

	reqHttp, err = http.NewRequestWithContext(ctx, http.MethodGet, "http://"+cs.cfg.Server.Addr+"/pics?search="+qe, nil)

	cs.Require().NoError(err)

	reqHttp.Header.Set("Authorization", h)
	reqHttp.Header.Set("Content-Type", "application/json")

	cs.T().Log("Sending GET request to pics endpoint...")
	resp, err = http.DefaultClient.Do(reqHttp)

	cs.Require().NoError(err)
	defer resp.Body.Close()

	resultPics := struct {
		URLs []string `json:"urls"`
	}{}

	dec := json.NewDecoder(resp.Body)

	err = dec.Decode(&resultPics)
	cs.Require().NoError(err)

	cs.Require().Equal(10, len(resultPics.URLs))

	for i, u := range resultPics.URLs {
		if u == "https://imgs.xkcd.com/comics/an_apple_a_day.png" {
			break
		}
		if i == 9 {
			cs.T().Error("expected comics not found")
		}
	}
}

func (cs *ComicsSuite) TestWrongPassword() {
	req := httpserver.LoginRequest{
		Username: "user1",
		Password: "wrong_password",
	}

	jsonData, err := json.Marshal(req)
	cs.Require().NoError(err)

	cs.T().Log("Running user test...")
	cs.T().Log("Sending POST request to login endpoint...")

	resp, err := http.Post("http://"+cs.cfg.Server.Addr+"/login", "application/json", bytes.NewBuffer(jsonData))
	cs.Require().NoError(err)

	defer resp.Body.Close()

	cs.Require().Equal(http.StatusUnauthorized, resp.StatusCode)

	h := resp.Header.Get("Authorization")

	cs.Require().True(len(h) == 0)
}

func (cs *ComicsSuite) TestUserNotExist() {
	req := httpserver.LoginRequest{
		Username: "not_exist",
		Password: "not_exist",
	}

	jsonData, err := json.Marshal(req)
	cs.Require().NoError(err)

	cs.T().Log("Running user test...")
	cs.T().Log("Sending POST request to login endpoint...")

	resp, err := http.Post("http://"+cs.cfg.Server.Addr+"/login", "application/json", bytes.NewBuffer(jsonData))
	cs.Require().NoError(err)

	defer resp.Body.Close()

	cs.Require().Equal(http.StatusUnauthorized, resp.StatusCode)

	h := resp.Header.Get("Authorization")

	cs.Require().True(len(h) == 0)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	reqHttp, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://"+cs.cfg.Server.Addr+"/update", nil)

	cs.Require().NoError(err)

	cs.T().Log("Sending POST request to update endpoint...")
	resp, err = http.DefaultClient.Do(reqHttp)

	cs.Require().Equal(http.StatusUnauthorized, resp.StatusCode)
	cs.Require().NoError(err)
	defer resp.Body.Close()

	qe := url.QueryEscape("apple doctor")

	reqHttp, err = http.NewRequestWithContext(ctx, http.MethodGet, "http://"+cs.cfg.Server.Addr+"/pics?search="+qe, nil)

	cs.Require().NoError(err)

	cs.T().Log("Sending GET request to pics endpoint...")
	resp, err = http.DefaultClient.Do(reqHttp)

	cs.Require().NoError(err)
	cs.Require().Equal(http.StatusUnauthorized, resp.StatusCode)
	defer resp.Body.Close()
}

func TestComicsSuite(t *testing.T) {
	s := new(ComicsSuite)
	defer s.TearDownSuite()

	suite.Run(t, s)
}
