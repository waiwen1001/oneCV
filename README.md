# Financial Assistance Scheme Management System Setup Guide

This guide explains how to set up the `Financial Assistance Scheme
Management System` backend, including configuring PostgreSQL with Docker and setting up Goose for database migrations.

---

## Prerequisites

Ensure the following tools are installed on your system:

- Docker (Recommended for PostgreSQL setup)
- Go (Golang) version 1.23.2
- Git (for version control)

---

## Installation Steps

### Step 1: Set Up PostgreSQL Using Docker
Run the following command to set up and start a PostgreSQL database container:
- Run the following command to download the PostgreSQL Docker image: 
```bash
docker pull postgres
```
- Start a PostgreSQL container with the required environment variables:
```bash
docker run --name onecv_db -e POSTGRES_USER=onecv_user -e POSTGRES_PASSWORD=onecv_pw -e POSTGRES_DB=onecv -p 5432:5432 -d postgres
```
### Step 2: Set Up Goose ( migrations ) Environment Variables
Run the following command to set up and goose migration environment:
- ( Linux/macOS )
```bash
export GOOSE_DRIVER="postgres"
export GOOSE_DBSTRING="postgresql://onecv_user:onecv_pw@localhost:5432/onecv"
export GOOSE_MIGRATION_DIR="./migrations"
```
- Window ( Powershell )
``` bash
$env:GOOSE_DRIVER = "postgres"
$env:GOOSE_DBSTRING = "postgresql://onecv_user:onecv_pw@localhost:5432/onecv"
$env:GOOSE_MIGRATION_DIR = "./migrations"
```
### Step 3: Set Up the `.env` File

Create a `.env` file in the root directory of your project and add the following environment variables:

```bash
DB_USER="onecv_user"
DB_PASSWORD="onecv_pw"
DB_NAME="onecv"
DB_CONNECTION="postgresql://onecv_user:onecv_pw@localhost:5432/onecv?sslmode=disable"
```

### Step 4: Install Dependencies and Run the Application

- **Install Dependencies:**  
   Run the following command to install the dependencies listed in your `go.mod` file:
   ```bash
   go get
   ```
- **Clean Up Dependencies:**  
   Use `go mod tidy` to remove any unused dependencies and ensure everything is clean:
   ```bash
   go mod tidy
   ```
- **Run the Application:**  
   Finally, run the application with the following command:
   ```bash
   go run main.go
   ```



## API Reference

#### Get all Applicants

```http
  GET /api/applicants
```
**Response**
- Success (200)
```bash
{
    "applicants": [
        {
            "id": "a02cb4d1-f98e-48ac-bc14-08f61749350c",
            "name": "Johnny",
            "employment_status": "employed",
            "sex": "male",
            "date_of_birth": "1985-06-21",
            "marital_status": "single",
            "household": [...]
        }
    ]
}
```

#### Create Applicant

```http
  POST /api/applicants
```
**Request body**

The request body should be a raw JSON object with the following fields:
``` bash
{
    "name": "John",
    "employment_status": "Unemployed",
    "sex": "Male",
    "date_of_birth": "1985-06-21",
    "marital_status": "Single",
    "household": [
        {
            "name": "Jake",
            "relation": "Son",
            "date_of_birth": "2018-02-11",
            "sex": "Male",
            "employment_status": "Unemployed"
        },
        {
            "name": "Micheal",
            "relation": "Daughter",
            "date_of_birth": "2004-11-22",
            "sex": "Female",
            "employment_status": "Unemployed"
        }
    ]
}
```

| Parameter            | Type     | Description                       |
| :--------            | :------- | :-------------------------------- |
| `name`               | `string` | **Required**. The name of the applicant |
| `employment_status`  | `string` | **Required**. The employment status of the applicant (e.g., "employed", "unemployed") |
| `sex`                | `string` | **Required**. The gender of the applicant |
| `date_of_birth`      | `string` | **Required**. The date of birth of the applicant (YYYY-MM-DD) |
| `marital_status`     | `string` | The marital status of the applicant (e.g., "single", "married", "widowed", "divorced") |
| `household`          | `array`  | A list of household members, each with the following fields: |
| - `name`             | `string` | **Required**. The name of the household member |
| - `relation`         | `string` | **Required**. The relation to the applicant (e.g., "Son", "Daughter") |
| - `date_of_birth`    | `string` | **Required**. The date of birth of the household member (YYYY-MM-DD) |
| - `sex`              | `string` | The gender of the household member |
| - `employment_status`| `string` | The employment status of the household member (e.g., "employed", "unemployed") |

**Response**
- Success (200)
```bash
{
    "message": "Applicant submitted successfully"
}
```

#### Get all Schemes

```http
  GET /api/schemes
```
**Response**
- Success (200)
```bash
{
    "schemes": [
        {
            "id": "0f30e79d-3cc2-4855-88f3-5ce33a42d9be",
            "name": "Retrenchment Assistance Scheme (families)",
            "criteria": {
                "employment_status": "unemployed",
                "has_children": {
                    "school_level": "== primary"
                }
            },
            "benefits": [...]
        }
    ]
}
```

#### Create new Scheme

```http
  POST /api/schemes
```
**Request body**

The request body should be a raw JSON object with the following fields:
``` bash
{
    "name": "Retrenchment Assistance Scheme (families)",
    "description": "Financial assistance for retrenched workers",
    "criteria": [
        {
            "conditions": {
                "employment_status": "unemployed"
            },
            "benefits": [
                {"name": "Benefit 001", "amount": 500.00},
                {"name": "Benefit 002", "amount": 100.00}
            ]
          },
          {
            "conditions": {
                "has_children": {
                    "school_level": "== primary"
                }
            },
            "benefits": [
                {"name": "Benefit 003", "amount": 1000.00},
                {"name": "Benefit 004", "amount": 50.00}
            ]
        }
    ]
}
```

| Parameter              | Type     | Description                       |
| :--------              | :------- | :-------------------------------- |
| `name`                 | `string` | **Required**. The name of the scheme |
| `description`          | `string` | **Required**. The description of the scheme  |
| `criteria`             | `array`  | **Required**. The criteria of the scheme |
| - `conditions`         | `object` | **Required**. Conditions to qualify for the benefit. |
| - `employment_status`  | `string` | **Required**. The employment status condition (e.g., "unemployed"). |
| - `has_children`       | `object` | **Required**. If children exist, specify conditions like `school_level`. |
| - `benefits`           | `array`  | **Required**. List of benefits provided for the criteria. |
| - `name`               | `string` | **Required**. The name of the benefit. |
| - `amount`             | `float`  | **Required**. The monetary value of the benefit. |

**Response**
- Success (200)
```bash
{
    "message": "Scheme submitted successfully"
}
```

#### Get Eligible Schemes for an Applicant

```http
  GET /api/schemes/eligible?applicant={id}
```
**Query Parameters**
| Parameter    | Type     | Description                       |
| :--------    | :------- | :-------------------------------- |
| `applicant`  | `string` | **Required.** The unique ID of the applicant.|

**Response**
- Success (200)
```bash
{
    "scheme": [
        {
            "id": "0f30e79d-3cc2-4855-88f3-5ce33a42d9be",
            "name": "Retrenchment Assistance Scheme (families)",
            "criteria": {
                "has_children": {
                    "school_level": "== primary"
                }
            },
            "benefits": [...]
        },
        {...}
    ]
}
```

#### Get All Applications
```http
  GET /api/applications
```

**Response**
- Success (200)
```bash
{
    "applications": [
        {
            "application_id": "398112eb-ba30-4c1f-a434-9a98c3755f01",
            "applicant": {
                "id": "a02cb4d1-f98e-48ac-bc14-08f61749350c",
                "name": "Johnny",
                "employment_status": "employed"
            },
            "scheme": {
                "id": "0f30e79d-3cc2-4855-88f3-5ce33a42d9be",
                "name": "Retrenchment Assistance Scheme (families)",
                "eligible": [...]
            },
            "status": "Pending",
            "submitted_at": "2025-01-11 19:51:30"
        },
        {...}
    ]
}
```


#### Get Applicant by ID
```http
  GET /api/applicants/{id}
```

**URL Parameters**
| Parameter   | Type     | Description                       |
| :--------   | :------- | :-------------------------------- |
| `id`        | `string` | **Required.** The unique ID of the applicant to get.|

**Response**
- Success (200)
```bash
{
    "applicant": {
        "id": "a02cb4d1-f98e-48ac-bc14-08f61749350c",
        "name": "Johnny",
        "employment_status": "employed",
        "sex": "male",
        "date_of_birth": "1985-06-21",
        "marital_status": "single",
        "household": [...]
    }
}
```

#### Create Application
```http
  POST /api/applications
```

**Request body**

The request body should be a raw JSON object with the following fields:
``` bash
{
    "applicant_id": "a02cb4d1-f98e-48ac-bc14-08f61749350c",
    "scheme_id": "0f30e79d-3cc2-4855-88f3-5ce33a42d9be"
}
```

**Response**
- Success (200)
```bash
{
    "message": "Application submitted successfully"
}
```

#### Update Applicant by Id
```http
  PUT /api/applicants/{id}
```

**URL Parameters**
| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `id`      | `string` | **Required.** The unique ID of the applicant to update.|

**Request body**

The request body should be a raw JSON object with the following fields:
``` bash
{
    "name": "Johnny",
    "employment_status": "Employed",
    "sex": "Male",
    "date_of_birth": "1985-06-21",
    "marital_status": "Single",
    "household": [
        {
            "name": "Jim",
            "relation": "Son",
            "date_of_birth": "2020-02-11",
            "sex": "Male",
            "employment_status": "Unemployed"
        },
        {
            "name": "Vicky",
            "relation": "Daughter",
            "date_of_birth": "2014-11-22",
            "sex": "Female",
            "employment_status": "Unemployed"
        }
    ]
}
```

**Response**
- Success (200)
```bash
{
    "message": "Applicant updated successfully"
}
```

#### Delete Applicant
```http
  DELETE /api/applicants/{id}
```

**URL Parameters**
| Parameter  | Type     | Description                       |
| :--------  | :------- | :-------------------------------- |
| `id`       | `string` | **Required.** The unique ID of the applicant to delete.|

**Response**
- Success (200)
```bash
{
    "message": "Applicant deleted successfully"
}
```

#### Update Scheme
```http
  PUT /api/schemes/{id}
```

**URL Parameters**
| Parameter  | Type     | Description                       |
| :--------  | :------- | :-------------------------------- |
| `id`       | `string` | **Required.** The unique ID of the applicant to update.|

**Request body**

The request body should be a raw JSON object with the following fields:
``` bash
{
    "name": "Retrenchment Assistance Scheme (families)",
    "description": "Financial assistance for retrenched workers",
    "criteria": [
        {
            "conditions": {
                "employment_status": "unemployed"
            },
            "benefits": [
                {"name": "Benefit 001", "amount": 500.00},
                {"name": "Benefit 002", "amount": 100.00}
            ]
        },
        {
            "conditions": {
                "has_children": {
                    "school_level": "== primary"
                }
            },
            "benefits": [
                {"name": "Benefit 003", "amount": 1000.00},
                {"name": "Benefit 004", "amount": 50.00}
            ]
        }
    ]
}
```

**Response**
- Success (200)
```bash
{
    "message": "Scheme upated successfully"
}
```

#### Delete Scheme
```http
  DELETE /api/schemes/{id}
```

**URL Parameters**
| Parameter   | Type     | Description                       |
| :--------   | :------- | :-------------------------------- |
| `id`        | `string` | **Required.** The unique ID of the scheme to delete.|

**Response**
- Success (200)
```bash
{
    "message": "Scheme deleted successfully"
}
```

#### Update Application
```http
  PUT /api/applications/{id}
```

**URL Parameters**
| Parameter   | Type     | Description                       |
| :--------   | :------- | :-------------------------------- |
| `id`        | `string` | **Required.** The unique ID of the application to update.|

**Request body**
```bash
{
    "status": "Approved"
}
```

**Response**
- Success (200)
```bash
{
    "message": "Application submitted successfully"
}
```

#### Delete Application
```http
  DELETE /api/applications/{id}
```

**URL Parameters**
| Parameter  | Type     | Description                       |
| :--------  | :------- | :-------------------------------- |
| `id`       | `string` | **Required.** The unique ID of the application to delete.|

**Response**
- Success (200)
```bash
{
    "message": "Application deleted successfully"
}
```



