
# Tests for Tezos Delegation Management Application in Go

## Table of Contents

1. [Introduction](#introduction)
2. [Dependencies](#dependencies)
3. [Setup](#setup)
4. [Tests](#tests)
    - [TestInitDB](#testinitdb)
    - [TestGetLastTimestamp](#testgetlasttimestamp)
    - [TestGetDelegations](#testgetdelegations)
    - [TestFetchDelegationsRealAPI](#testfetchdelegationsrealapi)
5. [Running the Tests](#running-the-tests)
6. [Conclusion](#conclusion)

## Introduction

This document provides information on how to run unit and integration tests for the Tezos Delegation  application built in Go. The tests are designed to verify the correct functionality of the application, ensuring that the core features work as expected.

## Dependencies

- **SQLite3**: The database for storing delegation data.
- **Gorilla Mux**: A HTTP router for Go.
- **Testify**: A toolkit with common assertions and mocks that plays nicely with the standard library.

To install these dependencies, use the following command:
```sh
go get -u github.com/gorilla/mux github.com/stretchr/testify
```

## Setup

Before running the tests, ensure that you have the necessary dependencies installed and the project is set up correctly.

1. **Install Dependencies**:
    ```sh
    go mod tidy
    ```

2. **Create the Test Database**:
    Ensure the SQLite3 database is initialized in memory for testing purposes. This is automatically handled by the `initTestDB` function in the test code.

## Tests

### TestInitDB

- **Purpose**: Verify that the database is correctly initialized.
- **Function**: `initTestDB()`
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

### TestFetchDelegationsRealAPI

- **Purpose**: Verify that the `fetchDelegationsFromURL` function retrieves and stores delegations from the real API.
- **Function**: `fetchDelegationsFromURL(db *sql.DB, url string, stopChan chan struct{})`
- **Assertions**:
  - At least one delegation should be inserted into the database.

## Running the Tests

To run the tests, use the following command in your terminal:

```sh
go test -v
```

## Conclusion

The tests ensure that the application's core functionalities are verified, including database initialization, data retrieval, and integration with the real API. Using an in-memory database ensures that tests are isolated and do not interfere with persistent data.