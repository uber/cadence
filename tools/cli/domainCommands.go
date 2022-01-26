// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/uber/cadence/tools/common/flag"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"

	"github.com/uber/cadence/client/frontend"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/domain"
	"github.com/uber/cadence/common/types"
)

var (
	gracefulFailoverType = "grace"
)

type (
	domainCLIImpl struct {
		// used when making RPC call to frontend service
		frontendClient frontend.Client

		// act as admin to modify domain in DB directly
		domainHandler domain.Handler
	}
)

// newDomainCLI creates a domain CLI
func newDomainCLI(
	c *cli.Context,
	isAdminMode bool,
) *domainCLIImpl {

	var frontendClient frontend.Client
	var domainHandler domain.Handler
	if !isAdminMode {
		frontendClient = initializeFrontendClient(c)
	} else {
		domainHandler = initializeAdminDomainHandler(c)
	}
	return &domainCLIImpl{
		frontendClient: frontendClient,
		domainHandler:  domainHandler,
	}
}

// RegisterDomain register a domain
func (d *domainCLIImpl) RegisterDomain(c *cli.Context) {
	domainName := getRequiredGlobalOption(c, FlagDomain)

	description := c.String(FlagDescription)
	ownerEmail := c.String(FlagOwnerEmail)
	retentionDays := defaultDomainRetentionDays

	if c.IsSet(FlagRetentionDays) {
		retentionDays = c.Int(FlagRetentionDays)
	}
	securityToken := c.String(FlagSecurityToken)
	var err error

	isGlobalDomain := true
	if c.IsSet(FlagIsGlobalDomain) {
		isGlobalDomain, err = strconv.ParseBool(c.String(FlagIsGlobalDomain))
		if err != nil {
			ErrorAndExit(fmt.Sprintf("Option %s format is invalid.", FlagIsGlobalDomain), err)
		}
	}

	var domainData *flag.StringMap
	if c.IsSet(FlagDomainData) {
		domainData = c.Generic(FlagDomainData).(*flag.StringMap)
	}
	if len(requiredDomainDataKeys) > 0 {
		err = checkRequiredDomainDataKVs(domainData.Value())
		if err != nil {
			ErrorAndExit("Domain data missed required data.", err)
		}
	}

	activeClusterName := ""
	if c.IsSet(FlagActiveClusterName) {
		activeClusterName = c.String(FlagActiveClusterName)
	}

	var clusters []*types.ClusterReplicationConfiguration
	if c.IsSet(FlagClusters) {
		clusterStr := c.String(FlagClusters)
		clusters = append(clusters, &types.ClusterReplicationConfiguration{
			ClusterName: clusterStr,
		})
		for _, clusterStr := range c.Args() {
			clusters = append(clusters, &types.ClusterReplicationConfiguration{
				ClusterName: clusterStr,
			})
		}
	}

	request := &types.RegisterDomainRequest{
		Name:                                   domainName,
		Description:                            description,
		OwnerEmail:                             ownerEmail,
		Data:                                   domainData.Value(),
		WorkflowExecutionRetentionPeriodInDays: int32(retentionDays),
		Clusters:                               clusters,
		ActiveClusterName:                      activeClusterName,
		SecurityToken:                          securityToken,
		HistoryArchivalStatus:                  archivalStatus(c, FlagHistoryArchivalStatus),
		HistoryArchivalURI:                     c.String(FlagHistoryArchivalURI),
		VisibilityArchivalStatus:               archivalStatus(c, FlagVisibilityArchivalStatus),
		VisibilityArchivalURI:                  c.String(FlagVisibilityArchivalURI),
		IsGlobalDomain:                         isGlobalDomain,
	}

	ctx, cancel := newContext(c)
	defer cancel()
	err = d.registerDomain(ctx, request)
	if err != nil {
		if _, ok := err.(*types.DomainAlreadyExistsError); !ok {
			ErrorAndExit("Register Domain operation failed.", err)
		} else {
			ErrorAndExit(fmt.Sprintf("Domain %s already registered.", domainName), err)
		}
	} else {
		fmt.Printf("Domain %s successfully registered.\n", domainName)
	}
}

// UpdateDomain updates a domain
func (d *domainCLIImpl) UpdateDomain(c *cli.Context) {
	domainName := getRequiredGlobalOption(c, FlagDomain)

	var updateRequest *types.UpdateDomainRequest
	ctx, cancel := newContext(c)
	defer cancel()

	if c.IsSet(FlagActiveClusterName) {
		activeCluster := c.String(FlagActiveClusterName)
		fmt.Printf("Will set active cluster name to: %s, other flag will be omitted.\n", activeCluster)

		var failoverTimeout *int32
		if c.String(FlagFailoverType) == gracefulFailoverType {
			timeout := int32(c.Int(FlagFailoverTimeout))
			failoverTimeout = &timeout
		}

		updateRequest = &types.UpdateDomainRequest{
			Name:                     domainName,
			ActiveClusterName:        common.StringPtr(activeCluster),
			FailoverTimeoutInSeconds: failoverTimeout,
		}
	} else {
		resp, err := d.describeDomain(ctx, &types.DescribeDomainRequest{
			Name: common.StringPtr(domainName),
		})
		if err != nil {
			if _, ok := err.(*types.EntityNotExistsError); !ok {
				ErrorAndExit("Operation UpdateDomain failed.", err)
			} else {
				ErrorAndExit(fmt.Sprintf("Domain %s does not exist.", domainName), err)
			}
			return
		}

		description := resp.DomainInfo.GetDescription()
		ownerEmail := resp.DomainInfo.GetOwnerEmail()
		retentionDays := resp.Configuration.GetWorkflowExecutionRetentionPeriodInDays()
		emitMetric := resp.Configuration.GetEmitMetric()
		var clusters []*types.ClusterReplicationConfiguration

		if c.IsSet(FlagDescription) {
			description = c.String(FlagDescription)
		}
		if c.IsSet(FlagOwnerEmail) {
			ownerEmail = c.String(FlagOwnerEmail)
		}
		var domainData *flag.StringMap
		if c.IsSet(FlagDomainData) {
			domainData = c.Generic(FlagDomainData).(*flag.StringMap)
		}
		if c.IsSet(FlagRetentionDays) {
			retentionDays = int32(c.Int(FlagRetentionDays))
		}
		if c.IsSet(FlagClusters) {
			clusterStr := c.String(FlagClusters)
			clusters = append(clusters, &types.ClusterReplicationConfiguration{
				ClusterName: clusterStr,
			})
			for _, clusterStr := range c.Args() {
				clusters = append(clusters, &types.ClusterReplicationConfiguration{
					ClusterName: clusterStr,
				})
			}
		}

		var binBinaries *types.BadBinaries
		if c.IsSet(FlagAddBadBinary) {
			if !c.IsSet(FlagReason) {
				ErrorAndExit("Must provide a reason.", nil)
			}
			binChecksum := c.String(FlagAddBadBinary)
			reason := c.String(FlagReason)
			operator := getCurrentUserFromEnv()
			binBinaries = &types.BadBinaries{
				Binaries: map[string]*types.BadBinaryInfo{
					binChecksum: {
						Reason:   reason,
						Operator: operator,
					},
				},
			}
		}

		var badBinaryToDelete *string
		if c.IsSet(FlagRemoveBadBinary) {
			badBinaryToDelete = common.StringPtr(c.String(FlagRemoveBadBinary))
		}

		updateRequest = &types.UpdateDomainRequest{
			Name:                                   domainName,
			Description:                            common.StringPtr(description),
			OwnerEmail:                             common.StringPtr(ownerEmail),
			Data:                                   domainData.Value(),
			WorkflowExecutionRetentionPeriodInDays: common.Int32Ptr(retentionDays),
			EmitMetric:                             common.BoolPtr(emitMetric),
			HistoryArchivalStatus:                  archivalStatus(c, FlagHistoryArchivalStatus),
			HistoryArchivalURI:                     common.StringPtr(c.String(FlagHistoryArchivalURI)),
			VisibilityArchivalStatus:               archivalStatus(c, FlagVisibilityArchivalStatus),
			VisibilityArchivalURI:                  common.StringPtr(c.String(FlagVisibilityArchivalURI)),
			BadBinaries:                            binBinaries,
			Clusters:                               clusters,
			DeleteBadBinary:                        badBinaryToDelete,
		}
	}

	securityToken := c.String(FlagSecurityToken)
	updateRequest.SecurityToken = securityToken
	_, err := d.updateDomain(ctx, updateRequest)
	if err != nil {
		if _, ok := err.(*types.EntityNotExistsError); !ok {
			ErrorAndExit("Operation UpdateDomain failed.", err)
		} else {
			ErrorAndExit(fmt.Sprintf("Domain %s does not exist.", domainName), err)
		}
	} else {
		fmt.Printf("Domain %s successfully updated.\n", domainName)
	}
}

func (d *domainCLIImpl) DeprecateDomain(c *cli.Context) {
	domainName := getRequiredGlobalOption(c, FlagDomain)
	securityToken := c.String(FlagSecurityToken)
	force := c.Bool(FlagForce)

	ctx, cancel := newContext(c)
	defer cancel()

	if !force {
		// check if there is any workflow in this domain, if exists, do not deprecate
		wfs, _ := listClosedWorkflow(getWorkflowClient(c), 1, 0, time.Now().UnixNano(), domainName, "", "", workflowStatusNotSet, nil, c)
		if len(wfs) > 0 {
			ErrorAndExit("Operation DeprecateDomain failed.", errors.New("workflow history not cleared in this domain"))
			return
		}
		wfs, _ = listOpenWorkflow(getWorkflowClient(c), 1, 0, time.Now().UnixNano(), domainName, "", "", nil, c)
		if len(wfs) > 0 {
			ErrorAndExit("Operation DeprecateDomain failed.", errors.New("workflow still running in this domain"))
			return
		}
	}
	err := d.deprecateDomain(ctx, &types.DeprecateDomainRequest{
		Name:          domainName,
		SecurityToken: securityToken,
	})
	if err != nil {
		if _, ok := err.(*types.EntityNotExistsError); !ok {
			ErrorAndExit("Operation DeprecateDomain failed.", err)
		} else {
			ErrorAndExit(fmt.Sprintf("Domain %s does not exist.", domainName), err)
		}
	} else {
		fmt.Printf("Domain %s successfully deprecated.\n", domainName)
	}
}

// FailoverDomains is used for managed failover all domains with domain data IsManagedByCadence=true
func (d *domainCLIImpl) FailoverDomains(c *cli.Context) {
	// ask user for confirmation
	prompt("You are trying to failover all managed domains, continue? Y/N")
	d.failoverDomains(c)
}

// return succeed and failed domains for testing purpose
func (d *domainCLIImpl) failoverDomains(c *cli.Context) ([]string, []string) {
	targetCluster := getRequiredOption(c, FlagActiveClusterName)
	domains := d.getAllDomains(c)
	shouldFailover := func(domain *types.DescribeDomainResponse) bool {
		isDomainNotActiveInTargetCluster := domain.ReplicationConfiguration.GetActiveClusterName() != targetCluster
		return isDomainNotActiveInTargetCluster && isDomainFailoverManagedByCadence(domain)
	}
	var succeedDomains []string
	var failedDomains []string
	for _, domain := range domains {
		if shouldFailover(domain) {
			domainName := domain.GetDomainInfo().GetName()
			err := d.failover(c, domainName, targetCluster)
			if err != nil {
				printError(fmt.Sprintf("Failed failover domain: %s\n", domainName), err)
				failedDomains = append(failedDomains, domainName)
			} else {
				fmt.Printf("Success failover domain: %s\n", domainName)
				succeedDomains = append(succeedDomains, domainName)
			}
		}
	}
	fmt.Printf("Succeed %d: %v\n", len(succeedDomains), succeedDomains)
	fmt.Printf("Failed  %d: %v\n", len(failedDomains), failedDomains)
	return succeedDomains, failedDomains
}

func (d *domainCLIImpl) getAllDomains(c *cli.Context) []*types.DescribeDomainResponse {
	var res []*types.DescribeDomainResponse
	pagesize := int32(200)
	var token []byte
	ctx, cancel := newContext(c)
	defer cancel()
	for more := true; more; more = len(token) > 0 {
		listRequest := &types.ListDomainsRequest{
			PageSize:      pagesize,
			NextPageToken: token,
		}
		listResp, err := d.listDomains(ctx, listRequest)
		if err != nil {
			ErrorAndExit("Error when list domains info", err)
		}
		token = listResp.GetNextPageToken()
		res = append(res, listResp.GetDomains()...)
	}
	return res
}

func isDomainFailoverManagedByCadence(domain *types.DescribeDomainResponse) bool {
	domainData := domain.DomainInfo.GetData()
	return strings.ToLower(strings.TrimSpace(domainData[common.DomainDataKeyForManagedFailover])) == "true"
}

func (d *domainCLIImpl) failover(c *cli.Context, domainName string, targetCluster string) error {
	updateRequest := &types.UpdateDomainRequest{
		Name:              domainName,
		ActiveClusterName: common.StringPtr(targetCluster),
	}
	ctx, cancel := newContext(c)
	defer cancel()
	_, err := d.updateDomain(ctx, updateRequest)
	return err
}

// DescribeDomain updates a domain
func (d *domainCLIImpl) DescribeDomain(c *cli.Context) {
	domainName := c.GlobalString(FlagDomain)
	domainID := c.String(FlagDomainID)
	printJSON := c.Bool(FlagPrintJSON)

	request := types.DescribeDomainRequest{}
	if domainID != "" {
		request.UUID = &domainID
	}
	if domainName != "" {
		request.Name = &domainName
	}
	if domainID == "" && domainName == "" {
		ErrorAndExit("At least domainID or domainName must be provided.", nil)
	}

	ctx, cancel := newContext(c)
	defer cancel()
	resp, err := d.describeDomain(ctx, &request)
	if err != nil {
		if _, ok := err.(*types.EntityNotExistsError); !ok {
			ErrorAndExit("Operation DescribeDomain failed.", err)
		}
		ErrorAndExit(fmt.Sprintf("Domain %s does not exist.", domainName), err)
	}

	if printJSON {
		output, err := json.Marshal(resp)
		if err != nil {
			ErrorAndExit("Failed to encode domain response into JSON.", err)
		}
		fmt.Println(string(output))
		return
	}

	clusters := "N/A, Not a global domain"
	if resp.IsGlobalDomain {
		clusters = clustersToString(resp.ReplicationConfiguration.Clusters)
	}
	var formatStr = "Name: %v\nUUID: %v\nDescription: %v\nOwnerEmail: %v\nDomainData: %v\nStatus: %v\nRetentionInDays: %v\n" +
		"EmitMetrics: %v\nIsGlobal(XDC)Domain: %v\nActiveClusterName: %v\nClusters: %v\nHistoryArchivalStatus: %v\n"
	descValues := []interface{}{
		resp.DomainInfo.GetName(),
		resp.DomainInfo.GetUUID(),
		resp.DomainInfo.GetDescription(),
		resp.DomainInfo.GetOwnerEmail(),
		resp.DomainInfo.Data,
		resp.DomainInfo.GetStatus(),
		resp.Configuration.GetWorkflowExecutionRetentionPeriodInDays(),
		resp.Configuration.GetEmitMetric(),
		resp.IsGlobalDomain,
		resp.ReplicationConfiguration.GetActiveClusterName(),
		clusters,
		resp.Configuration.GetHistoryArchivalStatus().String(),
	}
	if resp.Configuration.GetHistoryArchivalURI() != "" {
		formatStr = formatStr + "HistoryArchivalURI: %v\n"
		descValues = append(descValues, resp.Configuration.GetHistoryArchivalURI())
	}
	formatStr = formatStr + "VisibilityArchivalStatus: %v\n"
	descValues = append(descValues, resp.Configuration.GetVisibilityArchivalStatus().String())
	if resp.Configuration.GetVisibilityArchivalURI() != "" {
		formatStr = formatStr + "VisibilityArchivalURI: %v\n"
		descValues = append(descValues, resp.Configuration.GetVisibilityArchivalURI())
	}
	fmt.Printf(formatStr, descValues...)
	if resp.Configuration.BadBinaries != nil {
		fmt.Println("Bad binaries to reset:")
		table := tablewriter.NewWriter(os.Stdout)
		table.SetBorder(true)
		table.SetColumnSeparator("|")
		header := []string{"Binary Checksum", "Operator", "Start Time", "Reason"}
		headerColor := []tablewriter.Colors{tableHeaderBlue, tableHeaderBlue, tableHeaderBlue, tableHeaderBlue}
		table.SetHeader(header)
		table.SetHeaderColor(headerColor...)
		for cs, bin := range resp.Configuration.BadBinaries.Binaries {
			row := []string{cs}
			row = append(row, bin.GetOperator())
			row = append(row, time.Unix(0, bin.GetCreatedTimeNano()).String())
			row = append(row, bin.GetReason())
			table.Append(row)
		}
		table.Render()
	}
	if resp.GetFailoverInfo() != nil {
		info := resp.GetFailoverInfo()
		fmt.Println("Graceful failover info:")
		table := tablewriter.NewWriter(os.Stdout)
		table.SetBorder(true)
		table.SetColumnSeparator("|")
		header := []string{"Failover Version", "Start Time", "Expire Time", "Completed Shard Count", "Pending Shard"}
		table.SetHeader(header)
		row := []string{}
		row = append(row, fmt.Sprintf("%v", info.GetFailoverVersion()))
		row = append(row, time.Unix(0, info.GetFailoverStartTimestamp()).String())
		row = append(row, time.Unix(0, info.GetFailoverExpireTimestamp()).String())
		row = append(row, fmt.Sprintf("%v", info.GetCompletedShardCount()))
		row = append(row, fmt.Sprintf("%v", info.GetPendingShards()))
		table.Append(row)
		table.Render()
	}
}

func (d *domainCLIImpl) ListDomains(c *cli.Context) {
	pageSize := c.Int(FlagPageSize)
	prefix := c.String(FlagPrefix)
	printAll := c.Bool(FlagAll)
	printDeprecated := c.Bool(FlagDeprecated)
	printFull := c.Bool(FlagPrintFullyDetail)
	printJSON := c.Bool(FlagPrintJSON)

	if printAll && printDeprecated {
		ErrorAndExit(fmt.Sprintf("Cannot specify %s and %s flags at the same time.", FlagAll, FlagDeprecated), nil)
	}

	domains := d.getAllDomains(c)
	var filteredDomains []*types.DescribeDomainResponse

	// Only list domains that are matching to the prefix if prefix is provided
	if len(prefix) > 0 {
		var prefixDomains []*types.DescribeDomainResponse
		for _, domain := range domains {
			if strings.Index(domain.DomainInfo.Name, prefix) == 0 {
				prefixDomains = append(prefixDomains, domain)
			}
		}
		domains = prefixDomains
	}

	if printAll {
		filteredDomains = domains
	} else {
		filteredDomains = make([]*types.DescribeDomainResponse, 0, len(domains))
		for _, domain := range domains {
			if printDeprecated && *domain.DomainInfo.Status == types.DomainStatusDeprecated {
				filteredDomains = append(filteredDomains, domain)
			} else if !printDeprecated && *domain.DomainInfo.Status == types.DomainStatusRegistered {
				filteredDomains = append(filteredDomains, domain)
			}
		}
	}

	if printJSON {
		output, err := json.Marshal(filteredDomains)
		if err != nil {
			ErrorAndExit("Failed to encode domain results into JSON.", err)
		}
		fmt.Println(string(output))
		return
	}

	table := createTableForListDomains(printAll, printFull)

	currentPageSize := 0
	for i, domain := range filteredDomains {
		appendDomainToTable(table, domain, printAll, printFull)
		currentPageSize++

		if currentPageSize != pageSize {
			continue
		}

		// page is full
		table.Render()
		if i == len(domains)-1 || !showNextPage() {
			return
		}
		table.ClearRows()
		currentPageSize = 0
	}

	table.Render()
}

func (d *domainCLIImpl) listDomains(
	ctx context.Context,
	request *types.ListDomainsRequest,
) (*types.ListDomainsResponse, error) {

	if d.frontendClient != nil {
		return d.frontendClient.ListDomains(ctx, request)
	}

	return d.domainHandler.ListDomains(ctx, request)
}

func (d *domainCLIImpl) registerDomain(
	ctx context.Context,
	request *types.RegisterDomainRequest,
) error {

	if d.frontendClient != nil {
		return d.frontendClient.RegisterDomain(ctx, request)
	}

	return d.domainHandler.RegisterDomain(ctx, request)
}

func (d *domainCLIImpl) updateDomain(
	ctx context.Context,
	request *types.UpdateDomainRequest,
) (*types.UpdateDomainResponse, error) {

	if d.frontendClient != nil {
		return d.frontendClient.UpdateDomain(ctx, request)
	}

	return d.domainHandler.UpdateDomain(ctx, request)
}

func (d *domainCLIImpl) deprecateDomain(
	ctx context.Context,
	request *types.DeprecateDomainRequest,
) error {

	if d.frontendClient != nil {
		return d.frontendClient.DeprecateDomain(ctx, request)
	}

	return d.domainHandler.DeprecateDomain(ctx, request)
}

func (d *domainCLIImpl) describeDomain(
	ctx context.Context,
	request *types.DescribeDomainRequest,
) (*types.DescribeDomainResponse, error) {

	if d.frontendClient != nil {
		return d.frontendClient.DescribeDomain(ctx, request)
	}

	return d.domainHandler.DescribeDomain(ctx, request)
}

func createTableForListDomains(printAll, printFull bool) *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetColumnSeparator("|")
	header := []string{"Name", "UUID", "Domain Data"}
	if printAll || printFull {
		header = append(header, "Status")
	}
	header = append(header, "Is Global Domain", "Active Cluster")
	if printFull {
		header = append(header, "Clusters", "Retention Days", "History Archival Status", "History Archival URI", "Visibility Archival Status", "Visibility Archival URI")
	}
	headerColor := make([]tablewriter.Colors, len(header))
	for i := range headerColor {
		headerColor[i] = tableHeaderBlue
	}
	table.SetHeader(header)
	table.SetHeaderColor(headerColor...)
	table.SetHeaderLine(false)

	return table
}

func appendDomainToTable(
	table *tablewriter.Table,
	domain *types.DescribeDomainResponse,
	printAll bool,
	printFull bool,
) {
	row := []string{
		domain.DomainInfo.GetName(),
		domain.DomainInfo.GetUUID(),
		mapToString(domain.DomainInfo.GetData(), ", "),
	}
	if printAll || printFull {
		row = append(row, domain.DomainInfo.GetStatus().String())
	}
	row = append(row, strconv.FormatBool(domain.GetIsGlobalDomain()), domain.ReplicationConfiguration.GetActiveClusterName())
	if printFull {
		row = append(row,
			clustersToString(domain.ReplicationConfiguration.GetClusters()),
			fmt.Sprintf("%v", domain.Configuration.GetWorkflowExecutionRetentionPeriodInDays()),
			domain.Configuration.GetHistoryArchivalStatus().String(),
			domain.Configuration.GetHistoryArchivalURI(),
			domain.Configuration.GetVisibilityArchivalStatus().String(),
			domain.Configuration.GetVisibilityArchivalURI(),
		)
	}
	table.Append(row)
}

func archivalStatus(c *cli.Context, statusFlagName string) *types.ArchivalStatus {
	if c.IsSet(statusFlagName) {
		switch c.String(statusFlagName) {
		case "disabled":
			return types.ArchivalStatusDisabled.Ptr()
		case "enabled":
			return types.ArchivalStatusEnabled.Ptr()
		default:
			ErrorAndExit(fmt.Sprintf("Option %s format is invalid.", statusFlagName), errors.New("invalid status, valid values are \"disabled\" and \"enabled\""))
		}
	}
	return nil
}

func clustersToString(clusters []*types.ClusterReplicationConfiguration) string {
	var res string
	for i, cluster := range clusters {
		if i == 0 {
			res = res + cluster.GetClusterName()
		} else {
			res = res + ", " + cluster.GetClusterName()
		}
	}
	return res
}
