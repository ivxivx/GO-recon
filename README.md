# IO
- resource: a wrapper for data (in the form of bytes or file) which resides in local machine or remotely
- reader: a reader reads data from resource and unmarshalls it according to certain format
- writer: a writer marshalls data into certain format and writes to resource
- transformer: there are a few types of transformers
  - record extractor: a record extractor extracts records from input, e.g. extracts Transactions from the response of Get Transactions API
  - field extractor: a field extract fields from input, e.g. extracts fields (amount, currency, etc.) from a transaction record
  - field transformer: a field transformer transforms a field (a cell in CSV) from one format to another

```mermaid
graph LR
   Resource --> Reader --> Transformer --> Processor("(Processor)") --> Writer --> Transformer --> Resource
```

- Processors are specific to business scenarios, currenctly there is no interface for them.

## Resource
A resource may reside locally or remotely and can be accessed via certain protocol, such as HTTP or FTP.

Currently supported resources include:
- memory resource: data in memory
- local resource: a file on local file system
- http resource: resource can be accessed via HTTP (GET)
- sftp resource: resource can be accessed via Sftp

## Readers
Currently supported readers include:
- csv reader
- json reader
- composite reader: ustilize multiple readers (may handle different data formats) to read data

## Writers
Currently supported writer include:
- csv writer

# Usage
## Example
- csv reader + local resource: read a CSV file from local file system
- csv reader + sftp resource: read a CSV file from remote file system via Sftp
- json reader + http resource: read JSON data from remote system via HTTP
- csv writer + memory resource: format a report as CSV and send it via email, without storing data in file system
