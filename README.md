# PostgreSQL Table Data Masker (pgtdm)

## Overview

The PostgreSQL Table Data Masker (pgtdm) is an open-source tool designed to anonymize sensitive data from PostgreSQL database exports while preserving the structure of the data. It is particularly useful for creating sanitized datasets for development, testing, or analytics purposes. The tool supports PostgreSQL `COPY` blocks, like those from pg_dump, and relies on configurable anonymization rules defined in a JSON file.

## Features

- **Flexible Anonymization Rules**: Specify which columns to anonymize and their anonymization strategies using a JSON configuration file.
- **Wide Range of Anonymization Strategies**: Includes email, phone, address, UUID, and more.
- **Efficient Processing**: Handles large datasets efficiently with buffered I/O.
- **Customizable Output**: Allows you to define input and output file paths.
- **Open Source**: Community-driven and available under an open-source license.

## Installation

### Prerequisites

- Go 1.19 or later

### Steps

1. Clone the repository:
   ```bash
   git clone https://github.com/alexmgriffiths/pgtdm.git
   cd pgtdm
   ```
2. Build the project:
   ```bash
   make build
   ```

## Usage

### Command-Line Arguments

| Flag       | Description                        | Default        |
| ---------- | ---------------------------------- | -------------- |
| `--input`  | Path to the input file (required). | None           |
| `--output` | Path to the output file.           | `./output.txt` |
| `--rules`  | Path to the JSON rules file.       | `./rules.json` |

### Example Command

```bash
./pgtdm --input=sample.sql --output=anonymized.sql --rules=rules.json
```

### JSON Rules File Format

Define the anonymization rules in a JSON file. Each key represents a table, and each value is an array of columns to anonymize. Specify replacement types using `:`. For example:

```json
{
    "public.users": [
        "email",
        "username",
        "phone",
        "address",
        "password",
        "faceid_key:uuid",
        "first_name",
        "last_name",
        "work_address_city:city",
        "work_address_state:state",
        "work_address_zip:zip"
    ],
    "public.devices": [
        "token:uuid"
    ]
}
```

### Understanding Column Masking Syntax

The column masking syntax supports two types of configurations:

1. **Implicit Masking**: When no colon (`:`) is provided, the tool assumes the column name directly corresponds to its anonymization type. For example, the column `email` will use the `email` anonymization strategy.

2. **Explicit Masking**: A colon (`:`) can be used to specify the exact masking strategy to be applied. This is helpful when a column name does not clearly indicate its data type. For instance, `faceid_key:uuid` explicitly specifies that the `faceid_key` column should be masked as a UUID.

#### Example:
- `email` will be masked as an email address (implicit).
- `token:uuid` will be masked as a UUID (explicit).
- `work_address_city:city` specifies that the `work_address_city` column should use the `city` masking strategy.

### Available Masking Strategies

| Masking Strategy | Description                                     |
| ----------------- | ----------------------------------------------- |
| `address`         | Generates a full address string.               |
| `username`        | Generates a random username.                   |
| `firstname`       | Generates a random first name.                 |
| `lastname`        | Generates a random last name.                  |
| `phone`           | Generates a random phone number.               |
| `street`          | Generates a random street address.             |
| `city`            | Generates a random city name.                  |
| `state`           | Generates a random state name.                 |
| `zip`             | Generates a random ZIP code.                   |
| `password`        | Generates a random password.                   |
| `pin`             | Generates a random 4-digit PIN.                |
| `ssn`             | Generates a random Social Security Number.     |
| `companyname`     | Generates a random company name.               |
| `ein`             | Generates a random Employer Identification Number (EIN). |
| `bankname`        | Generates a random bank name.                  |
| `uuid`            | Generates a random UUID.                       |
| `ip`              | Generates a random IP address.                 |
| `email`           | Generates a random email address.              |

## Development

### Code Structure

- **`main.go`**: Entry point for the application.
- **`internal/anonymizer`**: Core logic for data anonymization.
- **`internal/random`**: Provides functions for generating random values (e.g., names, emails, UUIDs).
- **`internal/file`**: Utilities for file handling.
- **`internal/logging`**: Logging utilities for error handling and debugging.

### Contributing

Contributions are welcome! If you have ideas for improvements or new features, feel free to open an issue or submit a pull request.

1. Fork the repository.
2. Create a feature branch.
3. Make your changes.
4. Submit a pull request.

## License

This project is licensed under the MIT License. See the `LICENSE` file for more details.

## Contact

For questions or support, please open an issue on the [GitHub repository](https://github.com/alexmgriffiths/pgtdm)