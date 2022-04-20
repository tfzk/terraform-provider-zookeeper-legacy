package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/tfzk/terraform-provider-zookeeper/internal/client"
)

const (
	providerFieldServers        = "servers"
	providerFieldSessionTimeout = "session_timeout"

	providerEnvFieldServers        = "ZOOKEEPER_SERVERS"
	providerEnvFieldSessionTimeout = "ZOOKEEPER_SESSION"
)

func New() (*schema.Provider, error) {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			providerFieldServers: {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   false,
				DefaultFunc: schema.EnvDefaultFunc(providerEnvFieldServers, nil),
				Description: "A string containing a comma separated list of 'host:port' pairs",
			},
			providerFieldSessionTimeout: {
				Type:        schema.TypeInt,
				Optional:    true,
				Sensitive:   false,
				DefaultFunc: schema.EnvDefaultFunc(providerEnvFieldSessionTimeout, 10),
				Description: "How many seconds a session is considered valid after losing connectivity",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			typeZNode:    resourceZNode(),
			typeSeqZNode: resourceSeqZNode(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			typeZNode: datasourceZNode(),
		},
		ConfigureContextFunc: configureProviderContext,
	}, nil
}

func configureProviderContext(ctx context.Context, rscData *schema.ResourceData) (interface{}, diag.Diagnostics) {
	servers := rscData.Get(providerFieldServers).(string)
	sessionTimeout := rscData.Get(providerFieldSessionTimeout).(int)

	diags := diag.Diagnostics{}

	if servers != "" {
		c, err := client.NewClient(servers, sessionTimeout)

		if err != nil {
			// Report inability to connect internal Client
			return nil, append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Unable creating ZooKeeper client against '%s': %v", servers, err),
			})
		}

		return c, diags
	}

	// Report missing mandatory arguments
	return nil, append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  fmt.Sprintf("Provider requires at least the '%s' argument", providerFieldServers),
	})
}