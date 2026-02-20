@wip
Feature: Network Error Handling
  As a Canopy user
  I want graceful handling of network errors
  So that I can work offline and sync when connected

  Background:
    Given Canopy is installed
    And cloud sync is configured

  @critical
  Scenario: Sync with no network connection
    Given the network is unavailable
    When I run "canopy sync"
    Then the command should fail gracefully
    And the error should mention network or connection
    And local memories should remain intact
    And memories should be queued in the outbox

  @critical
  Scenario: Network timeout during sync
    Given the network is slow (>30s timeout)
    When I run "canopy sync"
    Then the command should timeout
    And the error should mention timeout
    And partial sync should not corrupt data

  Scenario: Network drops mid-sync
    Given I am syncing 100 memories
    And the network drops after 50 memories
    When the sync fails
    Then 50 memories should be synced
    And 50 memories should remain in outbox
    And no data should be lost

  Scenario: DNS resolution failure
    Given DNS is not resolving
    When I run "canopy sync"
    Then the command should fail
    And the error should mention connection or DNS

  Scenario: SSL certificate error
    Given the server has an invalid SSL certificate
    When I run "canopy sync"
    Then the command should fail
    And the error should mention certificate or TLS

  @critical
  Scenario: Server returns 401 Unauthorized
    Given my API key is invalid or expired
    When I run "canopy sync"
    Then the command should fail
    And the error should mention "authentication" or "unauthorized"
    And the error should suggest checking API key

  Scenario: Server returns 403 Forbidden
    Given my account is suspended
    When I run "canopy sync"
    Then the command should fail
    And the error should mention "forbidden" or "access denied"

  Scenario: Server returns 429 Rate Limited
    Given I have exceeded the rate limit
    When I run "canopy sync"
    Then the command should fail
    And the error should mention "rate limit"
    And the error should suggest waiting

  Scenario: Server returns 500 Internal Error
    Given the server is experiencing errors
    When I run "canopy sync"
    Then the command should fail
    And the error should mention "server error"
    And the command should suggest retrying later

  Scenario: Server returns 503 Service Unavailable
    Given the server is under maintenance
    When I run "canopy sync"
    Then the command should fail
    And the error should mention "unavailable" or "maintenance"

  @critical
  Scenario: Offline-first operation
    Given the network is unavailable
    When I run "canopy remember 'offline memory'"
    Then the command should succeed
    And the memory should be stored locally
    And the memory should be queued for sync

  Scenario: Outbox drain on reconnect
    Given I have 10 memories in the sync outbox
    And the network was previously unavailable
    When the network becomes available
    And the outbox drainer runs
    Then all 10 memories should be synced
    And the outbox should be empty

  Scenario: Exponential backoff on repeated failures
    Given sync has failed 3 times
    When the outbox drainer retries
    Then it should wait with exponential backoff
    And the wait time should include jitter
    And it should not hammer the server

  Scenario: Crown API health check
    When I run "canopy sync-status"
    Then the command should check Crown API health
    And the output should show connection status
