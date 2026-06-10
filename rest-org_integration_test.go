package sdk_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	sdk "github.com/kubermatic/grafanasdk"
)

func deleteOrgIfExists(t *testing.T, client *sdk.Client, ctx context.Context, oName string) {
	t.Helper()
	org, err := client.GetOrgByOrgName(ctx, oName)
	if err != nil {
		if errors.As(err, &sdk.ErrNotFound{}) {
			return
		}
		t.Fatalf("failed to lookup org %q for cleanup: %s", oName, err)
	}
	_, err = client.DeleteOrg(ctx, org.ID)
	if err != nil && !errors.As(err, &sdk.ErrNotFound{}) {
		t.Fatalf("failed to delete org %q for cleanup: %s", oName, err)
	}
}

func TestCreateDelete(t *testing.T) {
	shouldSkip(t)

	client := getClient(t)
	ctx := context.Background()

	oName := "coolorg"
	deleteOrgIfExists(t, client, ctx, oName)

	o := sdk.Org{Name: oName}
	statusMessage, err := client.CreateOrg(ctx, o)
	if err != nil {
		t.Fatalf("failed to create an org: %v (%s)", statusMessage, err.Error())
	}

	oID := *statusMessage.OrgID

	retrievedOrg, err := client.GetOrgById(ctx, oID)
	if err != nil {
		t.Fatalf("failed to retrieved org: %s", err.Error())
	}

	if retrievedOrg.Name != o.Name {
		t.Fatalf("got wrong org: got %s, expected %s", retrievedOrg.Name, o.Name)
	}

	_, err = client.DeleteOrg(ctx, oID)
	if err != nil {
		t.Fatalf("failed to delete org: %s", err.Error())
	}

	_, err = client.GetOrgById(ctx, oID)
	if err == nil {
		t.Fatalf("org %s is still there even though it should be deleted", o.Name)
	}
	if !errors.As(err, &sdk.ErrNotFound{}) {
		t.Fatalf("expected ErrNotFound, got: %s", err.Error())
	}
}

// TestUpdateOrgAddress checks if updating Org address works correctly
func TestUpdateOrgAddress(t *testing.T) {
	shouldSkip(t)

	client := getClient(t)
	ctx := context.Background()

	oName := "coolorg"
	deleteOrgIfExists(t, client, ctx, oName)

	// Create a new organization
	o := sdk.Org{Name: oName}
	statusMessage, err := client.CreateOrg(ctx, o)
	if err != nil {
		t.Fatalf("failed to create an org: %v (%s)", statusMessage, err.Error())
	}
	oID := *statusMessage.OrgID
	t.Cleanup(func() {
		// switch back to org 1 before deleting, otherwise subsequent tests
		// run in a stale/deleted org context and get 403s
		client.SwitchActualUserContext(ctx, 1)
		_, err := client.DeleteOrg(ctx, oID)
		if err != nil && !errors.As(err, &sdk.ErrNotFound{}) {
			t.Errorf("failed to cleanup org %d: %s", oID, err)
		}
	})

	// Test if updating organization by ID works as expected
	// Create a dummy address object
	address := sdk.Address{
		Address1: "CoolAddress1",
		Address2: "CoolAddress2",
		City:     "CoolCity",
		ZipCode:  "CoolZipCode",
		State:    "CoolState",
		Country:  "CoolCountry",
	}

	// Try updating organization address by Org ID
	statusMessage, err = client.UpdateOrgAddress(ctx, address, oID)
	if err != nil {
		t.Fatalf("failed to update the address: %v (%s)", statusMessage, err.Error())
	}

	retrievedOrg, err := client.GetOrgById(ctx, oID)
	if err != nil {
		t.Fatalf("failed to retrieved org: %s", err.Error())
	}

	// Check if retrieved address values are equal to expected ones
	if !reflect.DeepEqual(retrievedOrg.Address, address) {
		t.Fatalf("got wrong address: got %+v, expected %+v", retrievedOrg.Address, address)
	}

	// Test if updating current organization works as expected
	address = sdk.Address{
		Address1: "NiceAddress1",
		Address2: "NiceAddress2",
		City:     "NiceCity",
		ZipCode:  "NiceZipCode",
		State:    "NiceState",
		Country:  "NiceCountry",
	}

	statusMessage, err = client.SwitchActualUserContext(ctx, oID)
	if err != nil {
		t.Fatalf("failed to switch user context: %s", err.Error())
	}

	// Try updating current organization address
	statusMessage, err = client.UpdateActualOrgAddress(ctx, address)
	if err != nil {
		t.Fatalf("failed to update the address: %v (%s)", statusMessage, err.Error())
	}

	retrievedOrg, err = client.GetActualOrg(ctx)
	if err != nil {
		t.Fatalf("failed to retrieved org: %s", err.Error())
	}

	// Check if retrieved address values are equal to expected ones
	if !reflect.DeepEqual(retrievedOrg.Address, address) {
		t.Fatalf("got wrong address: got %v, expected %v", retrievedOrg.Address, address)
	}
}
