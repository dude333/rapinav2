# googlesheets

A Go package that provides a high-level, interface-driven API for creating and manipulating Google Sheets documents. Authentication uses OAuth 2.0 with a "Desktop app" credential so that **only you** can access the sheets the program creates. The OAuth scope is restricted to `drive.file`, which means the application can only see files it creates itself — it cannot read or list any other file in your Google Drive.

---

## Interface

```go
type Spreadsheet interface {
    NewSheet(sheetName string) error
    SetZoom(zoomScale float64) error
    SetColWidth(widths []float64)
    FreezePane(cell string) error
    SetFont(size float64, bold, wrap bool) (int, error)
    SetNumber(size float64, bold bool, format string) (int, error)
    PrintCell(row, col, style int, value any)
    RemoveRow(row int) error
    RemoveCol(col int) error
    SaveAs(name string) error
    Close() error
}
```

> **Note on `SetZoom`:** The Google Sheets REST API v4 does not expose a zoom-level field via `batchUpdate`. The method exists in the interface for forward-compatibility and is currently a no-op.

---

## Step-by-step: Configuring Google Cloud access

Follow these steps once to obtain a `credentials.json` file that the package uses to authenticate on your behalf.

### Step 1 — Create a Google Cloud project

1. Open the [Google Cloud Console](https://console.cloud.google.com/).
2. Click the project selector at the top of the page, then **New Project**.
3. Give it a name (e.g. `my-sheets-app`) and click **Create**.
4. Make sure the new project is selected in the project selector before proceeding.

### Step 2 — Enable the required APIs

You need to enable two APIs: **Google Sheets API** and **Google Drive API**.

1. In the left menu go to **APIs & Services → Library**.
2. Search for **Google Sheets API**, click it, then click **Enable**.
3. Go back to the Library, search for **Google Drive API**, click it, then click **Enable**.

### Step 3 — Configure the OAuth consent screen

> This step defines what the user sees when the application asks for permission. Because only you will ever use this application you can keep the app in _Testing_ mode indefinitely.

1. Go to **APIs & Services → OAuth consent screen**.
2. Choose **External** as the user type and click **Create**.
3. Fill in the required fields:
   - **App name** — any name, e.g. `my-sheets-app`
   - **User support email** — your email address
   - **Developer contact information** → your email address
4. Click **Save and Continue** on all remaining screens until you reach **Summary**, then click **Back to Dashboard**.
5. On the consent screen dashboard, click **Publish App** — or leave it in _Testing_ mode and add yourself as a test user in the next sub-step.

   **If staying in Testing mode:** Click **Add Users** under _Test users_ and add the Google account you will authenticate with. Only listed users can complete the OAuth flow while the app is in testing mode.

### Step 4 — Create an OAuth 2.0 client credential

1. Go to **APIs & Services → Credentials**.
2. Click **Create Credentials → OAuth client ID**.
3. For _Application type_ choose **Desktop app**.
4. Give it a name (e.g. `googlesheets-cli`) and click **Create**.
5. In the dialog that appears click **Download JSON**.
6. Rename the downloaded file to `credentials.json` and place it in the working directory from which you run your program (or pass its path via `Config.CredentialsFile`).

> **Keep `credentials.json` private.** Add it to your `.gitignore` and never commit it to version control.

### Step 5 — First run (interactive login)

On the first execution the package will print a URL like:

```
Authorise this application by visiting:

  https://accounts.google.com/o/oauth2/auth?...

Paste the authorisation code here and press Enter:
```

1. Open the URL in your browser.
2. Choose the Google account that should own the sheets.
3. Review the permissions ("See, edit, create, and delete only the specific Google Drive files you use with this app") and click **Allow**.
4. Copy the authorisation code shown in the browser.
5. Paste it into the terminal and press Enter.

The package will cache the resulting token in `token.json` (or `Config.TokenFile`). Subsequent runs load the cached token automatically and are fully non-interactive.

> **Keep `token.json` private.** It grants access on your behalf. Add it to `.gitignore`.

### Recommended `.gitignore` additions

```
credentials.json
token.json
```

---

## Installation

```bash
go get github.com/yourusername/googlesheets
```

> Replace `github.com/yourusername/googlesheets` with the actual module path you set in `go.mod`.

---

## Quick start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/yourusername/googlesheets"
)

func main() {
    ctx := context.Background()

    // Creates a new spreadsheet in your Google Drive.
    // New returns *Sheet, not an interface — you get the full concrete type.
    ss, err := googlesheets.New(ctx, googlesheets.Config{
        CredentialsFile: "credentials.json",
        TokenFile:       "token.json",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer ss.Close()

    // Register styles.
    header, _ := ss.SetFont(12, true, false)
    body, _ := ss.SetFont(11, false, false)
    money, _ := ss.SetNumber(11, false, `"$"#,##0.00`)

    // Layout.
    ss.SetColWidth([]float64{200, 120, 120, 160})
    ss.FreezePane("A2") // freeze the header row

    // Header.
    ss.PrintCell(0, 0, header, "Description")
    ss.PrintCell(0, 1, header, "Date")
    ss.PrintCell(0, 2, header, "Category")
    ss.PrintCell(0, 3, header, "Amount")

    // Data.
    ss.PrintCell(1, 0, body, "Office supplies")
    ss.PrintCell(1, 1, body, "2025-01-10")
    ss.PrintCell(1, 2, body, "Admin")
    ss.PrintCell(1, 3, money, 45.99)

    // Rename the spreadsheet and flush pending writes.
    if err := ss.SaveAs("Expense Report"); err != nil {
        log.Fatal(err)
    }

    // SpreadsheetURL() is a method on *Sheet — no type assertion needed.
    fmt.Println("Done:", ss.SpreadsheetURL())
}
```

---

## API reference

### `New(ctx, cfg) (*Sheet, error)`

Creates a new, blank Google Sheets spreadsheet and returns a `*Sheet`.

### `Open(ctx, cfg, spreadsheetID) (*Sheet, error)`

Wraps an existing spreadsheet. The `spreadsheetID` is the long alphanumeric string in the sheet's URL:
`https://docs.google.com/spreadsheets/d/<SPREADSHEET_ID>/edit`.

### `Config`

| Field             | Default            | Description                                |
| ----------------- | ------------------ | ------------------------------------------ |
| `CredentialsFile` | `credentials.json` | Path to the OAuth2 client-secret JSON file |
| `TokenFile`       | `token.json`       | Path to the cached OAuth2 token            |

### `Spreadsheet` methods

| Method                              | Description                                                                                              |
| ----------------------------------- | -------------------------------------------------------------------------------------------------------- |
| `NewSheet(name)`                    | Create a new tab and make it active                                                                      |
| `SetZoom(scale)`                    | No-op (API limitation)                                                                                   |
| `SetColWidth(widths)`               | Set column widths in pixels (index 0 = column A)                                                         |
| `FreezePane(cell)`                  | Freeze rows/columns up to cell in A1 notation                                                            |
| `SetFont(size, bold, wrap)`         | Register a text style; returns style index                                                               |
| `SetNumber(size, bold, fmt)`        | Register a numeric style; returns style index                                                            |
| `PrintCell(row, col, style, value)` | Write a value with a style (0-indexed)                                                                   |
| `RemoveRow(row)`                    | Delete a row (0-indexed)                                                                                 |
| `RemoveCol(col)`                    | Delete a column (0-indexed)                                                                              |
| _Warnings_                          | flush may now log a warning instead of failing when a delete request would remove all non-frozen columns |
| `SaveAs(name)`                      | Rename the spreadsheet and flush pending writes                                                          |
| `Close()`                           | Flush pending writes                                                                                     |

---

## Security notes

- The OAuth scope used is `https://www.googleapis.com/auth/drive.file`. This is the **most restrictive Drive scope** that still allows file creation. The application cannot read, list, or modify any file it did not create.
- `credentials.json` identifies your Cloud project; `token.json` grants runtime access. Treat both as secrets.
- Token expiry is handled automatically by the OAuth2 library (refresh tokens are long-lived for Desktop app credentials).

---

## Running the tests

```bash
go test ./...
```

The tests cover the A1-notation parser and config defaults and do not require network access.
