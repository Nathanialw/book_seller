#!/bin/bash
set -euo pipefail

# ---------------------------
# Configuration Section
# ---------------------------

# File paths
CONFIG_FILE="./config.txt"
MODEL_DIR="./internal/models"
STATE_FILE="./migrations/schema_state.json"
MIGRATION_DIR="./migrations"
HISTORY_DIR="$MIGRATION_DIR/history"
VERSION_PREFIX_LENGTH=5

# Database Configuration
DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="ecommerce"
DB_USER="admin"
DB_PASSWORD="securepassword"
DB_SSLMODE="disable"

# ---------------------------
# Initialization
# ---------------------------

declare -gA add_sql_lines=()    # Global associative array for ADD SQL
declare -gA drop_sql_lines=()   # Global associative array for DROP SQL
declare -gA new_state=()        # Global associative array for schema state

mkdir -p "$HISTORY_DIR"
mkdir -p "$MIGRATION_DIR"

if [[ ! -f "$STATE_FILE" ]]; then
  echo "{}" > "$STATE_FILE"
fi

# Lock to prevent concurrent runs
exec 9>/tmp/migrate.lock
flock -n 9 || { echo "[ERROR] Another migration process is running." >&2; exit 1; }

# ---------------------------
# Argument Parsing
# ---------------------------

ROLLBACK=false
VERBOSE=false
DRY_RUN=false
TARGET_VERSION=""

while getopts "rvdt:" opt; do
  case $opt in
    r) ROLLBACK=true ;;
    v) VERBOSE=true ;;
    d) DRY_RUN=true ;;
    t) TARGET_VERSION="$OPTARG" ;;
    *) exit 1 ;;
  esac
done

# ---------------------------
# Helper Functions
# ---------------------------

log() {
  if [[ "$VERBOSE" == true ]]; then
    echo "[INFO] $1" >&2
  fi
}

error_exit() {
  echo "[ERROR] $1" >&2
  exit 1
}

backup_file() {
  local file=$1
  if [[ -f "$file" ]]; then
    log "Backing up $file to ${file}.bak"
    cp "$file" "${file}.bak"
  fi
}

safe_sed() {
  local file=$1
  local pattern=$2
  local temp_file
  temp_file=$(mktemp)
  
  # Use cat instead of mv to avoid permission issues
  sed "$pattern" "$file" > "$temp_file" && cat "$temp_file" > "$file"
  rm -f "$temp_file"
}

join_by() {
  local sep="$1"; shift
  local out=""
  for arg; do
    [[ -z "$out" ]] && out="$arg" || out="${out}${sep}${arg}"
  done
  echo "$out"
}

write_migration_file() {
  local filename="$1"
  local content="$2"
  local fullpath="$MIGRATION_DIR/$filename"
  echo "$content" > "$fullpath"
  echo "$fullpath"
}

# ---------------------------
# Database Operations
# ---------------------------

execute_sql() {
  local sql_file=$1
  local direction=$2
  
  echo "=== Executing $direction migration: $sql_file ===" >&2
  cat "$sql_file" >&2
  
  if [[ "$DRY_RUN" == true ]]; then
    echo "[DRY RUN] Would execute $direction migration: $sql_file" >&2
    return 0
  fi

  # Use a temporary file to capture output
  local temp_out=$(mktemp)
  
  # Execute via sudo to completely bypass password prompt
  sudo -u postgres psql -d "$DB_NAME" -f "$sql_file" > "$temp_out" 2>&1
  
  # Check for errors
  if grep -q "ERROR" "$temp_out"; then
    echo "[ERROR] Migration failed. Full output:" >&2
    cat "$temp_out" >&2
    rm -f "$temp_out"
    error_exit "Migration failed"
  fi
  
  # Clean up
  rm -f "$temp_out"
  return 0
}

# ---------------------------
# Schema Parsing
# ---------------------------

parse_struct_fields() {
    local gofile="$1"
    local structname="$2"
    local inside=0
    local fields=()

    while IFS= read -r line; do
        # Skip comments and empty lines
        [[ "$line" =~ ^[[:space:]]*// ]] && continue
        [[ -z "$line" ]] && continue

        # Detect struct start
        if [[ "$line" =~ ^type[[:space:]]+$structname[[:space:]]+struct[[:space:]]*\{ ]]; then
            inside=1
            continue
        fi

        # Detect struct end
        if [[ "$inside" -eq 1 && "$line" =~ ^\}[[:space:]]*$ ]]; then
            break
        fi

        if [[ "$inside" -eq 1 ]]; then
            # Clean the line
            line=$(echo "$line" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')

            # Skip lines without a type (embedded structs)
            if [[ "$line" =~ \. ]]; then  # Only skip lines with dots
                echo "[DEBUG] Skipping embedded struct: $line" >&2
                continue
            fi

            # Parse field name and type
            read -ra tokens <<< "$line"
            if [[ ${#tokens[@]} -lt 2 ]]; then
                continue
            fi

            field_name="${tokens[0]}"
            field_type="${tokens[1]}"

            # Strip tags
            field_type=$(echo "$field_type" | sed 's/`.*`//;s/\[.*\]//')

            # Map Go types to SQL types
            case "$field_type" in
                string) sql_type="VARCHAR(255)" ;;
                int|int32|int64) sql_type="INTEGER" ;;
                float32) sql_type="FLOAT" ;;
                float64) sql_type="DOUBLE PRECISION" ;;
                bool) sql_type="BOOLEAN" ;;
                time.Time) sql_type="TIMESTAMP" ;;
                *) sql_type="TEXT" ;;
            esac

            fields+=("${field_name}:${sql_type}")
            echo "[DEBUG] Found field: $field_name ($field_type) -> $sql_type" >&2
        fi
    done < "$gofile"

    printf '%s\n' "${fields[@]}"
}

# ---------------------------
# Migration Generation
# ---------------------------

generate_migration_files() {
  local version_prefix=$1
  local prev_version_prefix=$2
  local changes_detected=false

  mapfile -t model_lines < <(awk '/^#begin models/{flag=1;next}/^#end models/{flag=0}flag' "$CONFIG_FILE")

  declare -A prev_fields_map
  local prev_state=$(cat "$HISTORY_DIR/schema_state_$prev_version_prefix.json")
  echo "$prev_state"

  # Read from .tables key
  for table in $(jq -r '.tables | keys[]' <<<"$prev_state"); do
    fields=$(jq -r --arg tbl "$table" '.tables[$tbl][]' <<<"$prev_state")
    prev_fields_map["$table"]=$(join_by "|" $fields)
  done

  for line in "${model_lines[@]}"; do
    read -r gofile structname tablename outfile <<< "$line"
    echo "[DEBUG] Processing $structname -> $tablename" >&2

    gofile_path="$MODEL_DIR/$gofile"
    [[ ! -f "$gofile_path" ]] && continue

    mapfile -t curr_fields < <(parse_struct_fields "$gofile_path" "$structname")
    IFS="|" read -r -a prev_fields_arr <<< "${prev_fields_map[$tablename]:-}"

    declare -A curr_field_map prev_field_map

    printf '[prev_fields_arr] entry: %q\n' "${prev_fields_arr[@]}"
    
    
    # Normalize keys to lowercase for comparison
    for f in "${curr_fields[@]}"; do
      key="${f%%:*}"
      curr_field_map["${key,,}"]="$f"
    done
    for f in "${prev_fields_arr[@]}"; do
      key="${f%%:*}"
      prev_field_map["${key,,}"]="$f"
    done

    # forward & undo maps
    declare -A forward_statements
    declare -A undo_statements

    # Added or changed fields
    for f in "${curr_fields[@]}"; do
      fname="${f%%:*}"
      ftype="${f#*:}"
      fname_lc="${fname,,}"

      if [[ "$fname" == *.* ]]; then
        echo "[DEBUG] Skipping embedded field $fname" >&2
        continue
      fi

      if [[ -z "${prev_field_map[$fname_lc]:-}" ]]; then
        # Field added
        changes_detected=true
        forward_statements["add_${fname_lc}"]="ALTER TABLE $tablename ADD COLUMN IF NOT EXISTS ${fname_lc} $ftype;"
        undo_statements["drop_${fname_lc}"]="ALTER TABLE $tablename DROP COLUMN IF EXISTS ${fname_lc} CASCADE;"
      else
        # Field exists, check if type changed
        prev_type="${prev_field_map[$fname_lc]#*:}"
        if [[ "$ftype" != "$prev_type" ]]; then
          changes_detected=true
          forward_statements["alter_type_${fname_lc}"]="ALTER TABLE $tablename ALTER COLUMN ${fname_lc} TYPE $ftype USING ${fname_lc}::${ftype};"
          undo_statements["alter_type_${fname_lc}"]="ALTER TABLE $tablename ALTER COLUMN ${fname_lc} TYPE $prev_type USING ${fname_lc}::${prev_type};"
        fi
      fi
    done

    # Removed fields
    for pf in "${prev_fields_arr[@]}"; do
      pfname="${pf%%:*}"
      pfname_lc="${pfname,,}"
      pftype="${pf#*:}"
      if [[ -z "${curr_field_map[$pfname_lc]:-}" ]]; then
        changes_detected=true
        forward_statements["drop_${pfname_lc}"]="ALTER TABLE $tablename DROP COLUMN IF EXISTS ${pfname_lc} CASCADE;"
        undo_statements["add_${pfname_lc}"]="ALTER TABLE $tablename ADD COLUMN IF NOT EXISTS ${pfname_lc} $pftype;"
      fi
    done

    local forward_sql="" undo_sql=""
    for stmt in "${forward_statements[@]}"; do forward_sql+="$stmt"$'\n'; done
    for stmt in "${undo_statements[@]}"; do undo_sql+="$stmt"$'\n'; done

    if [[ -n "$forward_sql" || -n "$undo_sql" ]]; then
      add_sql_lines["${version_prefix}_${outfile}"]="$forward_sql"
      drop_sql_lines["${version_prefix}_${outfile%.sql}_undo.sql"]="$undo_sql"
    fi

    new_state["$tablename"]=$(join_by "|" "${curr_fields[@]}")
  done

  $changes_detected && return 0 || return 1
}



# ---------------------------
# State Management
# ---------------------------

update_state() {
  local version_num=$1
  local -n _new_state=$2

  history_file="${HISTORY_DIR}/schema_state_${version_num}.json"
  log "Saving history snapshot for version $version_num to $history_file"
  
  # Build complete state JSON with both meta and tables
  local json_obj='{"meta": {'
  
  # Include existing metadata or initialize
  if jq -e '.meta' "$STATE_FILE" >/dev/null 2>&1; then
    json_obj+="$(jq -c '.meta' "$STATE_FILE" | sed 's/^{//; s/}$//')"
  else
    json_obj+='"rollback_count": 0'
  fi
  
  json_obj+='}, "tables": {'
  
  # Add tables data
  for tbl in "${!_new_state[@]}"; do
    IFS="|" read -r -a fields_arr <<< "${_new_state[$tbl]}"
    json_obj+="\"$tbl\":["
    local first=true
    for f in "${fields_arr[@]}"; do
      [[ -z "$f" ]] && continue
      if $first; then
        first=false
      else
        json_obj+=","
      fi
      json_obj+="\"$f\""
    done
    json_obj+="],"
  done
  json_obj="${json_obj%,}}}"

  if [[ "$DRY_RUN" == false ]]; then
    # Save to history file
    echo "$json_obj" > "$history_file"
    
    # Update current state file
    echo "$json_obj" > "$STATE_FILE"
  fi
}

# ---------------------------
# Rollback Implementation
# ---------------------------

rollback_migration() {
  local target_version=${1:-"previous"}
  local current_version
  current_version=$(grep -E '^#version [0-9]+' "$CONFIG_FILE" | awk '{print $2}' || echo "0")

  if [[ "$target_version" == "previous" ]]; then
    target_version=$((current_version - 1))
    [[ "$target_version" -lt 0 ]] && target_version=0
  fi

  if ! [[ "$target_version" =~ ^[0-9]+$ ]]; then
    error_exit "Invalid target version: $target_version"
  fi

  if [[ "$target_version" -ge "$current_version" ]]; then
    echo "Current version is $current_version, cannot roll forward to $target_version" >&2
    return 1
  fi

  echo "Rolling back from version $current_version to $target_version" >&2

  # Load current state for verification
  declare -A current_state
  local current_state_json=$(cat "$STATE_FILE")
  for table in $(jq -r 'keys[]' <<<"$current_state_json"); do
    current_state["$table"]=$(jq -r --arg tbl "$table" '.[$tbl][]' <<<"$current_state_json")
  done

  local current_rollback_count=$(jq -r '.meta.rollback_count // 0' "$STATE_FILE")
  local new_rollback_count=$((current_rollback_count + 1))
  update_state_rollback_count "$new_rollback_count"

  for ((v=current_version; v>target_version; v--)); do
    local version_prefix=$(printf "%0${VERSION_PREFIX_LENGTH}d" "$v")
    
    # Find all undo files for this version
    shopt -s nullglob
    local undo_files=("${MIGRATION_DIR}/${version_prefix}_migrations_undo.sql")
    shopt -u nullglob

    if (( ${#undo_files[@]} > 0 )); then
      for file in "${undo_files[@]}"; do
        if [[ -f "$file" ]]; then
          echo "Executing rollback: $file"
          
          if ! execute_sql "$file" "down"; then
            error_exit "Failed to execute rollback script: $file"
          fi
          
          # Archive with global rollback count prefix
          archive_migration_files "$v" "$new_rollback_count"
        else
          error_exit "Undo file not found: $file"
        fi
      done
    else
      error_exit "No rollback files found for version $v"
    fi
  done

  # Update state file after successful rollback
  if [[ "$DRY_RUN" == false ]]; then
    # Get current rollback count before we overwrite the state
    local current_rollback_count=$(jq -r '.meta.rollback_count // 0' "$STATE_FILE")
    
    if [[ "$target_version" -eq 0 ]]; then
      # For version 0, we want a clean slate but keep the rollback count
      echo "{\"meta\": {\"rollback_count\": $current_rollback_count}, \"tables\": {}}" > "$STATE_FILE"
    else
      # Restore the exact state from history
      local history_file="${HISTORY_DIR}/schema_state_${target_version}.json"
      if [[ -f "$history_file" ]]; then
        # Update just the rollback_count in the historical state
        jq --argjson count "$current_rollback_count" '.meta.rollback_count = $count' "$history_file" > "$STATE_FILE"
      else
        error_exit "History file not found for version $target_version: $history_file"
      fi
    fi

    # Update config version
    safe_sed "$CONFIG_FILE" "s/^#version .*/#version $target_version/"
    
    # Verify final state matches what we expect
    verify_rollback_state "$target_version"
    echo "Successfully rolled back to version $target_version" >&2
    echo "Rollback count: $current_rollback_count" >&2
  fi
}

# ---------------------------
# Archive Migrations
# ---------------------------

update_state_rollback_count() {
  local count=$1
  local tmp_file=$(mktemp)
  jq --argjson count "$count" '.meta.rollback_count = $count' "$STATE_FILE" > "$tmp_file"
  mv "$tmp_file" "$STATE_FILE"
}

archive_migration_files() {
  local version=$1
  local rollback_count=$2
  local version_prefix=$(printf "%0${VERSION_PREFIX_LENGTH}d" "$version")
  
  mkdir -p "${MIGRATION_DIR}/archived"
  
  # Archive with global rollback count prefix
  for file in "${MIGRATION_DIR}/${version_prefix}"_*.sql; do
    local base_name=$(basename "$file")
    mv "$file" "${MIGRATION_DIR}/archived/${rollback_count}_${base_name}"
  done

  mv "${HISTORY_DIR}/schema_state_${version_prefix}.json" "${MIGRATION_DIR}/archived/${rollback_count}_schema_state_${version_prefix}.json"

  echo "Archived version $version_prefix files with global rollback count $rollback_count"
}

# ---------------------------
# Verification Functions
# ---------------------------

verify_rollback_state() {
  local target_version=$1
  
  # Verify config version
  local current_version=$(grep -E '^#version [0-9]+' "$CONFIG_FILE" | awk '{print $2}')
  if [[ "$current_version" != "$target_version" ]]; then
    error_exit "Config version mismatch after rollback (expected $target_version, got $current_version)"
  fi

  # Verify state file
  if [[ "$target_version" -eq 0 ]]; then
    if [[ "$(jq -e 'length > 0' "$STATE_FILE")" == "true" ]]; then
      error_exit "State file not empty after rollback to version 0"
    fi
  else
    local history_file="${HISTORY_DIR}/schema_state_${target_version}.json"
    if ! diff -q "$STATE_FILE" "$history_file" >/dev/null; then
      error_exit "State file does not match history for version $target_version"
    fi
  fi
}

# ---------------------------
# Main Execution Flow
# ---------------------------

main() {
  if [[ "$ROLLBACK" == true ]]; then
    rollback_migration "$TARGET_VERSION"
    exit 0
  fi

  # Initialize state file structure if empty
  if [[ ! -s "$STATE_FILE" ]]; then
    echo '{"meta": {"rollback_count": 0}, "tables": {}}' > "$STATE_FILE"
  elif ! jq -e '.meta.rollback_count' "$STATE_FILE" >/dev/null; then
    # Migrate old state format to new format
    jq '.meta.rollback_count = 0' "$STATE_FILE" > "${STATE_FILE}.new"
    mv "${STATE_FILE}.new" "$STATE_FILE"
  fi

  version_num=$(grep -E '^#version [0-9]+' "$CONFIG_FILE" | awk '{print $2}' || echo "0")
  prev_version_prefix=$(printf "%0${VERSION_PREFIX_LENGTH}d" "$version_num")
  version_num=$((version_num + 1))
  version_prefix=$(printf "%0${VERSION_PREFIX_LENGTH}d" "$version_num")

  # Clear arrays before generating
  add_sql_lines=()
  drop_sql_lines=()
  new_state=()

  if generate_migration_files "$version_prefix" "$prev_version_prefix"; then
    changes_detected=true
  else
    changes_detected=false
  fi

  if [[ "$changes_detected" == true ]]; then
    # Write all migration files
    for outfile in "${!add_sql_lines[@]}"; do
      fullpath="${MIGRATION_DIR}/${outfile}"
      {
        echo "-- Migration Version: $version_prefix"
        echo "${add_sql_lines[$outfile]}"
      } > "$fullpath"
      log "Created migration file: $fullpath"
    done

    # Write all undo files
    for outfile in "${!drop_sql_lines[@]}"; do
      fullpath="${MIGRATION_DIR}/${outfile}"
      {
        echo "-- Undo Migration Version: $version_prefix"
        echo "${drop_sql_lines[$outfile]}"
      } > "$fullpath"
      log "Created undo file: $fullpath"
    done

    # Execute migrations with cleaner output
    echo "=== Applying Migrations ==="
    for outfile in "${!add_sql_lines[@]}"; do
      fullpath="${MIGRATION_DIR}/${outfile}"
      echo "Applying: ${outfile}"
      execute_sql "$fullpath" "up"
      echo "----------------------------------------"
    done

    update_state "$version_prefix" new_state
    safe_sed "$CONFIG_FILE" "s/^#version .*/#version $version_num/"

    echo ""
    echo "Migration complete. Version $version_prefix successfully applied."
    echo "Created:"
    for f in "${!add_sql_lines[@]}"; do echo "  - ${MIGRATION_DIR}/${f}"; done
    for f in "${!drop_sql_lines[@]}"; do echo "  - ${MIGRATION_DIR}/${f}"; done
  else
    echo "No schema changes detected. No migrations applied."
  fi
}

main