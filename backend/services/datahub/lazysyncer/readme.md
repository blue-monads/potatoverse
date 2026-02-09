# Lazysyncer

Lazysyncer is a service design to capture data changes from a local SQLite database and sync them to a remote "buddy" node. It handles both historical data and real-time incremental updates using a CDC (Change Data Capture) mechanism.

## Flow

1.  **SelfCDC (Local Capture)**:
    -   Triggers (`INSERT`, `UPDATE`, `DELETE`) are installed on local tables.
    -   Changes are captured into a side-table (`<table>__cdc`).
    -   Only the fact that a row changed is stored (Record ID and Operation). The actual row data is read from the source table during synchronization.

2.  **BuddyCDC (Remote Sync)**:
    -   Connects to a remote buddy transport.
    -   **Discovery**: Fetches metadata about remote tables.
    -   **Initialization**: Creates local storage tables (`zz_buddy_<id>`) for new remote tables.
    -   **Synchronization**:
        -   **Serial Sync**: Fetches historical records until the local state caches up to the start of the CDC log.
        -   **CDC Sync**: Fetches incremental updates from the remote CDC log.
    -   **Persistence**: Updates local metadata to track `SyncedRowID` and `SyncedCDCID`.

## Usage

The service runs an event loop that periodically checks for updates from the buddy and applies them to the local `zz_buddy_*` tables. Because it syncs data lazily without capturing a full snapshot of the records, the synced data may not reflect a strict point-in-time state of the system; instead, it represents an eventually converging state.
