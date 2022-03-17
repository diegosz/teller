package providers

import (
	"context"
	"errors"
	"fmt"
	"os"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/spectralops/teller/pkg/core"
)

var (
	ErrCloudFlareSourceFieldIsMissing = errors.New("`source` field is missing")
)

type CloudflareSecretsClient interface {
	SetWorkersSecret(ctx context.Context, script string, req *cloudflare.WorkersPutSecretRequest) (cloudflare.WorkersPutSecretResponse, error)
	DeleteWorkersSecret(ctx context.Context, script, secretName string) (cloudflare.Response, error)
}

type CloudflareSecrets struct {
	client CloudflareSecretsClient
}

func NewCloudflareSecretsClient() (core.Provider, error) {
	api, err := cloudflare.New(
		os.Getenv("CLOUDFLARE_API_KEY"),
		os.Getenv("CLOUDFLARE_API_EMAIL"),
	)

	if err != nil {
		return nil, err
	}

	cloudflare.UsingAccount(os.Getenv("CLOUDFLARE_ACCOUNT_ID"))(api) //nolint
	return &CloudflareSecrets{client: api}, nil
}

func (c *CloudflareSecrets) Name() string {
	return "cloudflare_workers_secret"
}

func (c *CloudflareSecrets) Put(p core.KeyPath, val string) error {

	if p.Source == "" {
		return ErrCloudFlareSourceFieldIsMissing
	}

	secretName, err := c.getSecretName(p)
	if err != nil {
		return err
	}

	secretRequest := cloudflare.WorkersPutSecretRequest{
		Name: secretName,
		Text: val,
		Type: cloudflare.WorkerSecretTextBindingType,
	}

	_, err = c.client.SetWorkersSecret(context.TODO(), p.Source, &secretRequest)

	return err
}

func (c *CloudflareSecrets) PutMapping(p core.KeyPath, m map[string]string) error {
	if p.Source == "" {
		return ErrCloudFlareSourceFieldIsMissing
	}

	for k, v := range m {
		ap := p.WithEnv(fmt.Sprintf("%v/%v", p.Path, k))

		err := c.Put(ap, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *CloudflareSecrets) Delete(p core.KeyPath) error {

	if p.Source == "" {
		return ErrCloudFlareSourceFieldIsMissing
	}

	secretName, err := c.getSecretName(p)
	if err != nil {
		return err
	}

	_, err = c.client.DeleteWorkersSecret(context.TODO(), p.Source, secretName)
	return err
}

func (c *CloudflareSecrets) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) {
	return nil, fmt.Errorf("%s does not support read functionality", c.Name())
}

func (c *CloudflareSecrets) Get(p core.KeyPath) (*core.EnvEntry, error) {
	return nil, fmt.Errorf("%s does not support read functionality", c.Name())
}

func (c *CloudflareSecrets) DeleteMapping(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement deleteMapping yet", c.Name())
}

func (c *CloudflareSecrets) getSecretName(p core.KeyPath) (string, error) {

	k := p.Field
	if k == "" {
		k = p.Env
	}
	if k == "" {
		return "", fmt.Errorf("key required for fetching secrets. Received \"\"")
	}
	return k, nil

}
