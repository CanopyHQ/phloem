package acceptance

import (
	"context"
	"os"
	"testing"

	"github.com/cucumber/godog"
)

// TestFeatures runs all Gherkin acceptance tests
func TestFeatures(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance tests in short mode")
	}

	tags := os.Getenv("GODOG_TAGS")
	if tags == "" {
		tags = "~@wip&&~@brew_gate"
	} else {
		tags = tags + "&&~@wip"
	}

	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t,
			Tags:     tags,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("acceptance tests failed")
	}
}

// TestSmokeFeatures runs only smoke tests (quick verification)
func TestSmokeFeatures(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance tests in short mode")
	}

	tags := os.Getenv("GODOG_TAGS")
	if tags == "" {
		tags = "@smoke&&~@wip&&~@brew_gate"
	} else {
		tags = tags + "&&~@wip"
	}

	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t,
			Tags:     tags,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("smoke tests failed")
	}
}

// TestCriticalFeatures runs critical path tests
func TestCriticalFeatures(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance tests in short mode")
	}

	tags := os.Getenv("GODOG_TAGS")
	if tags == "" {
		tags = "@critical&&~@wip&&~@brew_gate"
	} else {
		tags = tags + "&&~@wip"
	}

	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t,
			Tags:     tags,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("critical tests failed")
	}
}

// InitializeScenario sets up step definitions
func InitializeScenario(ctx *godog.ScenarioContext) {
	tc := &TestContext{
		ctx: context.Background(),
	}

	// MCP Server steps
	ctx.Step(`^the Phloem MCP server is running$`, tc.mcpServerRunning)
	ctx.Step(`^I send an initialize request to the MCP server$`, tc.sendMCPInitialize)
	ctx.Step(`^I should receive a valid initialization response$`, tc.checkValidInitResponse)
	ctx.Step(`^the response should contain protocol version "([^"]*)"$`, tc.checkProtocolVersion)
	ctx.Step(`^the response should contain server name "([^"]*)"$`, tc.checkServerName)
	ctx.Step(`^I request the list of available MCP tools$`, tc.requestToolsList)
	ctx.Step(`^I should receive a list containing "([^"]*)"$`, tc.checkListContains)
	ctx.Step(`^I call the MCP tool "([^"]*)"$`, tc.callMCPTool)
	ctx.Step(`^I call the MCP tool "([^"]*)" with content "([^"]*)"$`, tc.callMCPToolWithContent)
	ctx.Step(`^I call the MCP tool "([^"]*)" with query "([^"]*)"$`, tc.callMCPToolWithQuery)
	ctx.Step(`^I should receive a success response$`, tc.checkSuccessResponse)
	ctx.Step(`^I should receive an error response$`, tc.checkErrorResponse)

	// Memory steps
	ctx.Step(`^the memory store is initialized$`, tc.memoryStoreInitialized)
	ctx.Step(`^I have stored a memory with content "([^"]*)"$`, tc.storeMemory)
	ctx.Step(`^I have stored (\d+) memories$`, tc.storeMultipleMemories)
	ctx.Step(`^the response should contain a memory ID$`, tc.checkMemoryID)
	ctx.Step(`^the results should contain "([^"]*)"$`, tc.checkResultsContain)

	// System steps
	ctx.Step(`^the Phloem system is initialized$`, tc.systemInitialized)

	// Additional steps for auto-ingestion and other features
	ctx.Step(`^the transcript watcher is configured$`, tc.transcriptWatcherConfigured)
	ctx.Step(`^the transcript watcher is running$`, tc.transcriptWatcherRunning)
	ctx.Step(`^new content is added to a transcript file$`, tc.newContentAdded)
	ctx.Step(`^the watcher should detect the change$`, tc.watcherDetectsChange)
	ctx.Step(`^the new content should be ingested$`, tc.newContentIngested)
	ctx.Step(`^a transcript with user message "([^"]*)"$`, tc.transcriptWithUserMessage)
	ctx.Step(`^the transcript is ingested$`, tc.transcriptIngested)
	ctx.Step(`^a memory should be created with role "([^"]*)"$`, tc.memoryCreatedWithRole)
	ctx.Step(`^the memory should contain "([^"]*)"$`, tc.memoryContains)
	ctx.Step(`^a transcript with assistant response about (.+)$`, tc.transcriptWithAssistantResponse)
	ctx.Step(`^the memory should be tagged with "([^"]*)"$`, tc.memoryTaggedWith)
	ctx.Step(`^I request the list of available MCP resources$`, tc.requestResourcesList)
	ctx.Step(`^I should receive a list containing "([^"]*)"$`, tc.checkListContains)
	ctx.Step(`^I read the MCP resource "([^"]*)"$`, tc.readMCPResource)
	ctx.Step(`^I should receive a list of recent memories$`, tc.receiveRecentMemories)
	ctx.Step(`^the response should be valid JSON$`, tc.responseValidJSON)
	ctx.Step(`^I should receive memory statistics$`, tc.receiveMemoryStats)
	ctx.Step(`^the response should contain total_memories$`, tc.responseContainsTotalMemories)
	ctx.Step(`^the response should contain database_size$`, tc.responseContainsDatabaseSize)
	ctx.Step(`^I have stored memories with various tags$`, tc.storedMemoriesWithTags)

	// Graft steps
	ctx.Step(`^I have stored (\d+) memories with tags "([^"]*)"$`, tc.storedMemoriesWithTagCount)
	ctx.Step(`^I export a graft with tags "([^"]*)" to "([^"]*)"$`, tc.exportGraft)
	ctx.Step(`^the graft file should be created$`, tc.graftFileCreated)
	ctx.Step(`^the graft should contain (\d+) memories$`, tc.graftContainsMemories)
	ctx.Step(`^the graft manifest should have name "([^"]*)"$`, tc.graftManifestName)
	ctx.Step(`^I have a graft file "([^"]*)" with (\d+) memories$`, tc.createTestGraft)
	ctx.Step(`^I import the graft file "([^"]*)"$`, tc.importGraft)
	ctx.Step(`^(\d+) memories should be imported$`, tc.memoriesImported)
	ctx.Step(`^the imported memories should have tag "([^"]*)"$`, tc.importedMemoriesTagged)
	ctx.Step(`^I inspect the graft file "([^"]*)"$`, tc.inspectGraft)
	ctx.Step(`^I should see the graft manifest$`, tc.seeGraftManifest)
	ctx.Step(`^the manifest should show (\d+) memories$`, tc.manifestShowsMemories)
	ctx.Step(`^no memories should be imported$`, tc.noMemoriesImported)
	ctx.Step(`^I have stored a memory with content "([^"]*)"$`, tc.storeMemory)
	ctx.Step(`^I have a graft file "([^"]*)" containing "([^"]*)"$`, tc.createGraftWithContent)
	ctx.Step(`^the duplicate memory should not be created$`, tc.duplicateNotCreated)
	ctx.Step(`^only unique memories should be imported$`, tc.onlyUniqueImported)
	ctx.Step(`^I have stored memories with citations$`, tc.storedMemoriesWithCitations)
	ctx.Step(`^I export a graft including citations$`, tc.exportGraftWithCitations)
	ctx.Step(`^the graft should contain citation data$`, tc.graftContainsCitations)
	ctx.Step(`^citations should be preserved on import$`, tc.citationsPreserved)
	ctx.Step(`^I have an invalid graft file "([^"]*)"$`, tc.createInvalidGraft)
	ctx.Step(`^I try to import "([^"]*)"$`, tc.tryImportGraft)
	ctx.Step(`^I should receive an error$`, tc.checkErrorResponse)
	ctx.Step(`^the error should indicate invalid format$`, tc.errorIndicatesInvalidFormat)

	// CLI steps (run canopy commands, assert exit code and output)
	ctx.Step(`^Phloem is installed$`, tc.phloemInstalled)
	ctx.Step(`^I run "([^"]*)"$`, tc.runCLICommand)
	ctx.Step(`^the command should succeed$`, tc.checkCommandSucceeded)
	ctx.Step(`^the command should fail$`, tc.checkCommandFailed)
	ctx.Step(`^the command should fail with exit code (\d+)$`, tc.checkCommandFailedWithExitCode)
	ctx.Step(`^the output should show "([^"]*)"$`, tc.outputShouldShow)
	ctx.Step(`^the output should contain "([^"]*)"$`, tc.outputShouldContain)
	ctx.Step(`^the output should show ([^"]+)$`, tc.outputShouldShow)
	ctx.Step(`^the error should contain "([^"]*)"$`, tc.errorShouldContain)
	ctx.Step(`^the error should mention ([^"]+)$`, tc.errorShouldContain)
	ctx.Step(`^the command should fail with "([^"]*)"$`, tc.checkCommandFailedWithMessage)
}

// Step implementations are in steps.go
