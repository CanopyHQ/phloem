@wip
Feature: All CLI Commands
  As a Canopy user
  I want all CLI commands to work correctly
  So that I can use the full feature set

  Background:
    Given Canopy is installed
    And the memory store is initialized

  # ============================================
  # CORE MCP COMMANDS
  # ============================================

  @smoke @critical
  Scenario: canopy serve - Start MCP server
    When I run "canopy serve" in background
    And I send an MCP initialize request
    Then I should receive a valid MCP response
    And the server should be running

  Scenario: canopy serve - Port already in use
    Given another process is using the MCP stdin/stdout
    When I run "canopy serve"
    Then the command should handle the conflict gracefully

  @critical
  Scenario: canopy watch - Monitor transcripts
    Given the Cursor transcripts directory exists
    When I run "canopy watch" in background
    And a new transcript file is created
    Then the transcript should be ingested automatically

  Scenario: canopy watch - No transcripts directory
    Given the Cursor transcripts directory does not exist
    When I run "canopy watch"
    Then the command should show a warning
    And the command should wait for the directory to be created

  @critical
  Scenario: canopy ingest - Import transcripts
    Given I have 5 transcript files from the last 24 hours
    When I run "canopy ingest"
    Then the command should succeed
    And 5 transcripts should be imported

  Scenario: canopy ingest - Custom duration
    Given I have transcript files from the last 7 days
    When I run "canopy ingest 7d"
    Then all transcripts from the last 7 days should be imported

  Scenario: canopy ingest - No transcripts found
    Given no transcript files exist
    When I run "canopy ingest"
    Then the command should succeed
    And the output should indicate 0 transcripts found

  Scenario: canopy ingest - Invalid duration format
    When I run "canopy ingest invalid"
    Then the command should fail
    And the error should mention invalid duration format

  # ============================================
  # MEMORY MANAGEMENT COMMANDS
  # ============================================

  @smoke @critical
  Scenario: canopy export - Export as JSON
    Given I have stored 10 memories
    When I run "canopy export json /tmp/memories.json"
    Then the command should succeed
    And the file "/tmp/memories.json" should exist
    And the file should contain valid JSON with 10 memories

  Scenario: canopy export - Export as Markdown
    Given I have stored 5 memories
    When I run "canopy export markdown /tmp/memories.md"
    Then the command should succeed
    And the file "/tmp/memories.md" should exist
    And the file should contain markdown formatted memories

  Scenario: canopy export - Invalid format
    When I run "canopy export xml /tmp/memories.xml"
    Then the command should fail
    And the error should mention unsupported format

  Scenario: canopy export - Permission denied
    When I run "canopy export json /root/memories.json"
    Then the command should fail
    And the error should mention permission denied

  @critical
  Scenario: canopy verify - Verify citations
    Given I have a memory with citations
    When I run "canopy verify <memory_id>"
    Then the command should succeed
    And the output should show citation verification results

  Scenario: canopy verify - Invalid memory ID
    When I run "canopy verify nonexistent123"
    Then the command should fail
    And the error should mention memory not found

  Scenario: canopy verify - Memory without citations
    Given I have a memory without citations
    When I run "canopy verify <memory_id>"
    Then the command should succeed
    And the output should indicate no citations to verify

  Scenario: canopy decay - Decay confidence scores
    Given I have memories with citations
    When I run "canopy decay"
    Then the command should succeed
    And citation confidence scores should be reduced based on age

  Scenario: canopy decay - Empty database
    Given no memories exist
    When I run "canopy decay"
    Then the command should succeed
    And the output should indicate nothing to decay

  # ============================================
  # CLOUD SYNC COMMANDS
  # ============================================

  @critical
  Scenario: canopy sync - Sync with cloud
    Given I have a valid API key configured
    And I have 10 local memories
    When I run "canopy sync"
    Then the command should succeed
    And memories should be uploaded to cloud

  Scenario: canopy sync - No API key
    Given no API key is configured
    When I run "canopy sync"
    Then the command should fail
    And the error should mention missing API key
    And the error should suggest how to configure it

  Scenario: canopy sync - Network error
    Given the network is unavailable
    When I run "canopy sync"
    Then the command should fail
    And the error should mention network error
    And local memories should remain intact

  Scenario: canopy sync - Invalid API key
    Given an invalid API key is configured
    When I run "canopy sync"
    Then the command should fail
    And the error should mention authentication failed

  @critical
  Scenario: canopy sync-status - Show sync status
    Given I have synced with cloud before
    When I run "canopy sync-status"
    Then the command should succeed
    And the output should show last sync time
    And the output should show pending uploads count

  Scenario: canopy sync-status - Never synced
    Given I have never synced with cloud
    When I run "canopy sync-status"
    Then the command should succeed
    And the output should indicate never synced

  # ============================================
  # GRAFT COMMANDS
  # ============================================

  @smoke @critical
  Scenario: canopy graft package - Create graft
    Given I have 5 memories tagged "architecture"
    When I run "canopy graft package architecture /tmp/arch.graft"
    Then the command should succeed
    And the file "/tmp/arch.graft" should exist

  @smoke @critical
  Scenario: canopy graft unpack - Import graft
    Given I have a valid graft file "/tmp/test.graft"
    When I run "canopy graft unpack /tmp/test.graft"
    Then the command should succeed
    And the memories should be imported

  Scenario: canopy graft inspect - View graft contents
    Given I have a valid graft file "/tmp/test.graft"
    When I run "canopy graft inspect /tmp/test.graft"
    Then the command should succeed
    And the output should show graft metadata
    And no memories should be imported

  Scenario: canopy graft - Invalid file
    When I run "canopy graft unpack /tmp/invalid.graft"
    Then the command should fail
    And the error should mention invalid graft format

  Scenario: canopy graft - File not found
    When I run "canopy graft unpack /tmp/nonexistent.graft"
    Then the command should fail
    And the error should mention file not found

  # ============================================
  # IMPORT COMMANDS
  # ============================================

  @critical
  Scenario: canopy import chatgpt - Import ChatGPT history
    Given I have a valid ChatGPT export file "conversations.json"
    When I run "canopy import chatgpt conversations.json"
    Then the command should succeed
    And conversations should be imported as memories

  Scenario: canopy import chatgpt - Invalid JSON
    Given I have an invalid JSON file "invalid.json"
    When I run "canopy import chatgpt invalid.json"
    Then the command should fail
    And the error should mention invalid JSON

  Scenario: canopy import chatgpt - Wrong format
    Given I have a JSON file with wrong structure
    When I run "canopy import chatgpt wrong.json"
    Then the command should fail
    And the error should mention unexpected format

  @critical
  Scenario: canopy import claude - Import Claude history
    Given I have a valid Claude export directory
    When I run "canopy import claude ./claude-export/"
    Then the command should succeed
    And conversations should be imported as memories

  Scenario: canopy import claude - Invalid directory
    When I run "canopy import claude /nonexistent/"
    Then the command should fail
    And the error should mention directory not found

  # ============================================
  # LICENSE COMMANDS
  # ============================================

  @smoke @critical
  Scenario: canopy license - Show license status (free)
    Given I am on the free tier
    When I run "canopy license"
    Then the command should succeed
    And the output should show "Tier: Free"
    And the output should show "Memory Window: 14 days"

  Scenario: canopy license - Show license status (pro)
    Given I have an active Pro license
    When I run "canopy license"
    Then the command should succeed
    And the output should show "Tier: Pro"
    And the output should show unlimited access

  @critical
  Scenario: canopy activate - Activate license key
    Given I have a valid license key
    When I run "canopy activate <license_key>"
    Then the command should succeed
    And the license should be activated
    And running "canopy license" should show Pro tier

  Scenario: canopy activate - Invalid license key
    When I run "canopy activate invalid_key_12345"
    Then the command should fail
    And the error should mention invalid license key

  Scenario: canopy activate - Expired license key
    Given I have an expired license key
    When I run "canopy activate <expired_key>"
    Then the command should fail
    And the error should mention license expired

  Scenario: canopy activate - Already activated
    Given I already have an active license
    When I run "canopy activate <new_key>"
    Then the command should succeed
    And the new license should replace the old one

  Scenario: canopy upgrade - Open upgrade page
    When I run "canopy upgrade"
    Then the command should attempt to open browser
    And the URL should be "https://canopyhq.io/upgrade"

  # ============================================
  # CHROME EXTENSION COMMANDS
  # ============================================

  Scenario: canopy native-messaging - Start native host
    When I run "canopy native-messaging" in background
    Then the native messaging host should start
    And it should accept Chrome extension connections

  Scenario: canopy install-native - Install manifest
    When I run "canopy install-native"
    Then the command should succeed
    And the native messaging manifest should be installed
    And the manifest should point to canopy binary

  Scenario: canopy install-native - Permission denied
    Given the Chrome native messaging directory is not writable
    When I run "canopy install-native"
    Then the command should fail
    And the error should mention permission denied

  # ============================================
  # SUPPORT COMMANDS
  # ============================================

  @smoke @critical
  Scenario: canopy version - Show version
    When I run "canopy version"
    Then the command should succeed
    And the output should match "canopy v\d+\.\d+\.\d+"

  Scenario: canopy help - Show help
    When I run "canopy help"
    Then the command should succeed
    And the output should list available commands

  Scenario: canopy help <command> - Show command help
    When I run "canopy help status"
    Then the command should succeed
    And the output should show status command usage

  Scenario: canopy report-bug - Collect system info
    When I run "canopy report-bug"
    Then the command should collect system information
    And the output should show collected data
    And the command should attempt to open GitHub issues

  Scenario: canopy update - Update via Homebrew
    Given Canopy was installed via Homebrew
    When I run "canopy update"
    Then the command should run "brew upgrade canopy"

  Scenario: canopy update - Not installed via Homebrew
    Given Canopy was installed manually
    When I run "canopy update"
    Then the command should show a warning
    And the output should suggest manual update steps

  # ============================================
  # TELEMETRY COMMANDS
  # ============================================

  Scenario: canopy telemetry status - Show telemetry status
    When I run "canopy telemetry status"
    Then the command should succeed
    And the output should show telemetry enabled/disabled state

  Scenario: canopy telemetry enable - Enable telemetry
    Given telemetry is disabled
    When I run "canopy telemetry enable"
    Then the command should succeed
    And telemetry should be enabled

  Scenario: canopy telemetry disable - Disable telemetry
    Given telemetry is enabled
    When I run "canopy telemetry disable"
    Then the command should succeed
    And telemetry should be disabled

  Scenario: canopy telemetry show - Show pending events
    Given there are pending telemetry events
    When I run "canopy telemetry show"
    Then the command should succeed
    And the output should list pending events

  Scenario: canopy telemetry export - Export telemetry log
    When I run "canopy telemetry export /tmp/telemetry.log"
    Then the command should succeed
    And the telemetry log should be exported

  # ============================================
  # ERROR HANDLING
  # ============================================

  @critical
  Scenario: Unknown command
    When I run "canopy unknowncommand"
    Then the command should fail
    And the error should mention unknown command
    And the error should suggest "canopy help"

  @critical
  Scenario: Unknown subcommand
    When I run "canopy graft unknownsub"
    Then the command should fail
    And the error should mention unknown subcommand

  @critical
  Scenario: Missing required argument
    When I run "canopy export"
    Then the command should fail
    And the error should mention missing argument

  Scenario: Invalid flag
    When I run "canopy status --invalid-flag"
    Then the command should fail
    And the error should mention unknown flag
