package obsync

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var i int

type sleepClient struct {
}

func (s sleepClient) Info(ctx context.Context) (interface{}, error) {
	return nil, nil
}

func (s sleepClient) Exists(ctx context.Context, s2 string) bool {
	return false
}

func (s sleepClient) Put(ctx context.Context, filePath, key string) error {
	i = i + 1

	done := make(chan bool)
	go func() {
		time.Sleep(time.Millisecond * 2)
		done <- true
	}()

	select {
	case <-done:
		fmt.Printf("%v\n", filePath)
		return nil
	case <-ctx.Done():
		fmt.Printf("%v\n", ctx)
		return fmt.Errorf("down")
	}
}

func init() {
	_ = RegisterBucketClientFunc("sleep", func(config BucketConfig) (BucketClient, error) {
		return &sleepClient{}, nil
	})
}

func TestRunner_Start(t *testing.T) {
	runner, err := NewRunner(RunnerConfig{
		LocalPath: ".",
		Threads:   10,
		Timeout:   time.Second,
		BucketConfigs: []BucketConfig{
			{
				Type: "sleep",
				Name: "sleep0",
			},
			{
				Type: "sleep",
				Name: "sleep1",
			},
		},
	})

	assert.NotNil(t, runner)

	numbers, err := runner.TasksByPath(".", nil)
	assert.NoError(t, err)

	i = 0
	err = runner.Start(context.Background())
	assert.NoError(t, err)
	assert.EqualValues(t, i, len(numbers)*2)
}

func TestNewRunner(t *testing.T) {

	_, err := NewRunner(RunnerConfig{
		LocalPath: ".",
		Threads:   10,
		BucketConfigs: []BucketConfig{
			{Type: "sleep", Name: "sleep0"},
		},
	})

	assert.NoError(t, err)

	_, err = NewRunner(RunnerConfig{
		LocalPath: "/dev/null",
		Threads:   10,
		BucketConfigs: []BucketConfig{
			{Type: "sleep", Name: "sleep0"},
		},
	})
	assert.Error(t, err)

	_, err = NewRunner(RunnerConfig{
		LocalPath: "/etc",
		Threads:   0,
		BucketConfigs: []BucketConfig{
			{Type: "sleep", Name: "sleep0"},
		},
	})
	assert.Error(t, err)

	_, err = NewRunner(RunnerConfig{
		LocalPath: "/etc",
		Threads:   10,
		BucketConfigs: []BucketConfig{
			{Type: "sleepxxx", Name: "sleep0"},
		},
	})
	assert.Error(t, err)
}
