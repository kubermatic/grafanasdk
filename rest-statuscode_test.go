package sdk_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	sdk "github.com/kubermatic/grafanasdk"
	"github.com/stretchr/testify/assert"
)

// the API methods below used to discard the HTTP status code, so failed
// deletions were reported as success; these tests pin the mapping of
// 404 to ErrNotFound and of other non-200 codes to plain errors
func TestStatusCodeMapping(t *testing.T) {
	methods := map[string]func(ctx context.Context, client *sdk.Client) error{
		"DeleteUser": func(ctx context.Context, client *sdk.Client) error {
			_, err := client.DeleteUser(ctx, 1)
			return err
		},
		"DeleteOrg": func(ctx context.Context, client *sdk.Client) error {
			_, err := client.DeleteOrg(ctx, 1)
			return err
		},
		"GetOrgById": func(ctx context.Context, client *sdk.Client) error {
			_, err := client.GetOrgById(ctx, 1)
			return err
		},
		"GetOrgByOrgName": func(ctx context.Context, client *sdk.Client) error {
			_, err := client.GetOrgByOrgName(ctx, "org")
			return err
		},
		"DeleteDashboardByUID": func(ctx context.Context, client *sdk.Client) error {
			_, err := client.DeleteDashboardByUID(ctx, "uid")
			return err
		},
		"DeleteDatasourceByUID": func(ctx context.Context, client *sdk.Client) error {
			_, err := client.DeleteDatasourceByUID(ctx, "uid")
			return err
		},
	}

	responses := []struct {
		name         string
		code         int
		body         string
		wantErr      bool
		wantNotFound bool
	}{
		{
			name: "200 returns no error",
			code: http.StatusOK,
			// the body satisfies both StatusMessage and Org unmarshalling
			body:         `{"message":"ok","id":1,"name":"org"}`,
			wantErr:      false,
			wantNotFound: false,
		},
		{
			name:         "404 maps to ErrNotFound",
			code:         http.StatusNotFound,
			body:         `{"message":"not found"}`,
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:         "500 returns a plain error",
			code:         http.StatusInternalServerError,
			body:         `{"message":"boom"}`,
			wantErr:      true,
			wantNotFound: false,
		},
	}

	for method, call := range methods {
		for _, resp := range responses {
			t.Run(method+" "+resp.name, func(t *testing.T) {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(resp.code)
					_, _ = w.Write([]byte(resp.body))
				}))
				defer ts.Close()

				client, err := sdk.NewClient(ts.URL, "admin:admin", ts.Client())
				assert.Nil(t, err)

				err = call(context.Background(), client)
				assert.Equal(t, resp.wantErr, err != nil, "unexpected error state: %v", err)
				assert.Equal(t, resp.wantNotFound, errors.As(err, &sdk.ErrNotFound{}), "unexpected ErrNotFound mapping: %v", err)
			})
		}
	}
}
