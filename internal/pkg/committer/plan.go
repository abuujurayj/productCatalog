package committer

import (
	"context"

	cloud_spanner "cloud.google.com/go/spanner"
	"github.com/Vektor-AI/commitplan"
	"github.com/Vektor-AI/commitplan/drivers/spanner"
)

// Committer wraps the Spanner client and commitplan driver.
// It implements the commitplan.Committer interface.
type Committer struct {
	client *cloud_spanner.Client
}

func NewCommitter(client *cloud_spanner.Client) *Committer {
	return &Committer{
		client: client,
	}
}

// Apply executes the collected mutations atomically using the Spanner driver.
func (c *Committer) Apply(ctx context.Context, plan *commitplan.Plan) error {
	// Create the Spanner-specific driver provided by commitplan
	driver := spanner.NewDriver(c.client)

	// Execute the plan. This happens in a single Spanner transaction.
	return driver.Execute(ctx, plan)
}