// +build acceptance lbs

package v1

import (
	"testing"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/acceptance/tools"
	"github.com/rackspace/gophercloud/pagination"
	"github.com/rackspace/gophercloud/rackspace/lb/v1/lbs"
	"github.com/rackspace/gophercloud/rackspace/lb/v1/vips"
	th "github.com/rackspace/gophercloud/testhelper"
)

func TestLBs(t *testing.T) {
	return
	client := setup(t)

	ids := createLB(t, client, 3)

	listLBProtocols(t, client)

	listLBAlgorithms(t, client)

	listLBs(t, client)

	getLB(t, client, ids[0])

	updateLB(t, client, ids[0])

	deleteLB(t, client, ids[0])

	batchDeleteLBs(t, client, ids[1:])
}

func createLB(t *testing.T, client *gophercloud.ServiceClient, count int) []int {
	ids := []int{}

	for i := 0; i < count; i++ {
		opts := lbs.CreateOpts{
			Name:     tools.RandomString("test_", 5),
			Port:     80,
			Protocol: "HTTP",
			VIPs: []vips.VIP{
				vips.VIP{Type: vips.PUBLIC},
			},
		}

		lb, err := lbs.Create(client, opts).Extract()
		th.AssertNoErr(t, err)

		t.Logf("Created LB %d - waiting for it to build...", lb.ID)
		waitForLB(client, lb.ID, lbs.ACTIVE)
		t.Logf("LB %d has reached ACTIVE state", lb.ID)

		ids = append(ids, lb.ID)
	}

	return ids
}

func waitForLB(client *gophercloud.ServiceClient, id int, state lbs.Status) {
	gophercloud.WaitFor(60, func() (bool, error) {
		lb, err := lbs.Get(client, id).Extract()
		if err != nil {
			return false, err
		}
		if lb.Status != state {
			return false, nil
		}
		return true, nil
	})
}

func listLBProtocols(t *testing.T, client *gophercloud.ServiceClient) {
	err := lbs.ListProtocols(client).EachPage(func(page pagination.Page) (bool, error) {
		pList, err := lbs.ExtractProtocols(page)
		th.AssertNoErr(t, err)

		for _, p := range pList {
			t.Logf("Listing protocol: Name [%s]", p.Name)
		}

		return true, nil
	})
	th.AssertNoErr(t, err)
}

func listLBAlgorithms(t *testing.T, client *gophercloud.ServiceClient) {
	err := lbs.ListAlgorithms(client).EachPage(func(page pagination.Page) (bool, error) {
		aList, err := lbs.ExtractAlgorithms(page)
		th.AssertNoErr(t, err)

		for _, a := range aList {
			t.Logf("Listing algorithm: Name [%s]", a.Name)
		}

		return true, nil
	})
	th.AssertNoErr(t, err)
}

func listLBs(t *testing.T, client *gophercloud.ServiceClient) {
	err := lbs.List(client, lbs.ListOpts{}).EachPage(func(page pagination.Page) (bool, error) {
		lbList, err := lbs.ExtractLBs(page)
		th.AssertNoErr(t, err)

		for _, lb := range lbList {
			t.Logf("Listing LB: ID [%d] Name [%s] Protocol [%s] Status [%s] Node count [%d] Port [%d]",
				lb.ID, lb.Name, lb.Protocol, lb.Status, lb.NodeCount, lb.Port)
		}

		return true, nil
	})

	th.AssertNoErr(t, err)
}

func getLB(t *testing.T, client *gophercloud.ServiceClient, id int) {
	lb, err := lbs.Get(client, id).Extract()
	th.AssertNoErr(t, err)
	t.Logf("Getting LB %d: Created [%s] VIPs [%#v] Logging [%#v] Persistence [%#v] SourceAddrs [%#v]",
		lb.ID, lb.Created, lb.VIPs, lb.ConnectionLogging, lb.SessionPersistence, lb.SourceAddrs)
}

func updateLB(t *testing.T, client *gophercloud.ServiceClient, id int) {
	opts := lbs.UpdateOpts{
		Name:          tools.RandomString("new_", 5),
		Protocol:      "TCP",
		HalfClosed:    gophercloud.Enabled,
		Algorithm:     "RANDOM",
		Port:          8080,
		Timeout:       100,
		HTTPSRedirect: gophercloud.Disabled,
	}

	err := lbs.Update(client, id, opts).ExtractErr()
	th.AssertNoErr(t, err)

	t.Logf("Updating LB %d - waiting for it to finish", id)
	waitForLB(client, id, lbs.ACTIVE)
	t.Logf("LB %d has reached ACTIVE state", id)
}

func deleteLB(t *testing.T, client *gophercloud.ServiceClient, id int) {
	err := lbs.Delete(client, id).ExtractErr()
	th.AssertNoErr(t, err)
	t.Logf("Deleted %d", id)
}

func batchDeleteLBs(t *testing.T, client *gophercloud.ServiceClient, ids []int) {
	err := lbs.BulkDelete(client, ids).ExtractErr()
	th.AssertNoErr(t, err)
	t.Logf("Deleted %s", intsToStr(ids))
}
