# Transaction Reconciliation
Transaction reconciliation is the process of comparing two sets of transactions from two parties in order to find matches and discrepancies.

## Flow
The core flow is illustrated below:

```mermaid
sequenceDiagram
  actor Cron
  participant Recon as Reconciler
  participant PTY1 as [Party1]
  participant PTY2 as [Party2]

  Cron -> Recon: Trigger recon
  Recon ->> PTY1: Retrieve transactions of Party1
  PTY1 -->> Recon: result
  Recon ->> PTY2: Retrieve transactions of Party2
  PTY2 -->> Recon: result

  loop Every transaction of Party1
    Recon ->> Recon: Filter transaction
    Recon ->> Recon: Find matching<br/>transaction from Party2
    alt Found
      Recon ->> Recon: Compare two transactions
      Recon ->> Recon: Mark as `matched` or<br/>`mismatched`
    else
        Recon ->> Recon: Mark as `party1 only`
    end
  end

  loop Every transaction of Party2
    Recon ->> Recon: Similar to previous loop
  end
```

## Concepts
- Party: Reconciliation involves two parties.
- Collection: A collection contains transactions fetched from two parties.
- Filter: A filter uses some criteria to filter out  transactions before they can be passed over for comparison. Criteria may be a time range or a collection of statuses.
- Comparator: A comparator compares two transactions from two parties, in order to find whether they are matching.
