
## Table of Contents

1. [Introduction](#introduction)
2. [Setup](#setup)
3. [Unit Tests](#unit-tests)
4. [Integration Tests](#integration-tests)
5. [Running the Tests](#running-the-tests)
6. [Logs and Details](#logs-and-details)

## Introduction

This document provides information on how to run unit and integration tests for the Tezos Delegation Management application built in Go. The tests are designed to verify the correct functionality of the application, ensuring that the core features work as expected.

## Setup

Before running the tests, ensure that you have the necessary dependencies installed and the project is set up correctly.

1. **Install Dependencies**:
    ```sh
    go mod tidy
    ```

2. **Create the Database**:
    Ensure the SQLite3 database `delegations.db` is initialized and the `delegations` table exists. This is automatically handled by the `initDB` function in the application.

## Unit Tests

Unit tests focus on testing individual functions in isolation to ensure they behave correctly.

### TestInitDB

- **Purpose**: Verify that the database is correctly initialized.
- **Function**: `initDB()`
- **Assertions**:
  - The database connection should be successfully established.

### TestGetLastTimestamp

- **Purpose**: Verify that the `getLastTimestamp` function retrieves the latest timestamp from the database.
- **Function**: `getLastTimestamp(db *sql.DB) string`
- **Assertions**:
  - The timestamp should match the latest record in the `delegations` table.

### TestGetDelegations

- **Purpose**: Verify that the `getDelegations` function retrieves delegations correctly.
- **Function**: `getDelegations(w http.ResponseWriter, r *http.Request, db *sql.DB)`
- **Assertions**:
  - The HTTP response should be 200 OK.
  - The response body should match the expected JSON format.

## Integration Tests

Integration tests focus on testing the complete workflow of the application, ensuring that different components work together correctly.

### TestIntegration

- **Purpose**: Verify the end-to-end functionality of the application.
- **Steps**:
  - Initialize the database.
  - Start the HTTP server.
  - Insert a test delegation record.
  - Perform a GET request to retrieve the delegations.
- **Assertions**:
  - The HTTP response should be 200 OK.
  - The response body should match the expected JSON format.

## Running the Tests

To run the tests, use the following command in your terminal:

```sh
go test -v
