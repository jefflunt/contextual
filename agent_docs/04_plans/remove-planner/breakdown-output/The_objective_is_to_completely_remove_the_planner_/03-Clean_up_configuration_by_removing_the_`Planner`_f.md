# Clean up configuration by removing the `Planner` field and related logic from `internal/config/config.go`, and strip all related planner keys from `contextual.config.example.yml`.

This task involves modifying `internal/config/config.go` to completely remove the `Planner` field from the main configuration struct, along with any related sub-structs, default value initializations, and validation logic specific to the planner.

Additionally, you must update the `contextual.config.example.yml` file to remove the corresponding planner YAML keys. This ensures that the example configuration accurately reflects the updated struct and prevents users from attempting to configure a feature that no longer exists. Ensure that the YAML file remains well-formatted after the deletions.
