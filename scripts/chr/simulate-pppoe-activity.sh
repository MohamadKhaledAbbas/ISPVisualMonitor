#!/bin/bash

# MikroTik PPPoE Activity Simulation Script
# This script simulates realistic PPPoE connection/disconnection events
# for testing the ISP Visual Monitor's session tracking capabilities

set -e

# Configuration
CHR_HOST="${CHR_HOST:-localhost}"
CHR_API_PORT="${CHR_API_PORT:-8758}"
CHR_USERNAME="${CHR_USERNAME:-admin}"
CHR_PASSWORD="${CHR_PASSWORD:-}"

# PPPoE test users
USERS=("user001" "user002" "user003" "user004" "user005" "user006" "user007" "user008" "user009" "user010")
PASSWORDS=("pass001" "pass002" "pass003" "pass004" "pass005" "pass006" "pass007" "pass008" "pass009" "pass010")

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored messages
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if routeros CLI is available
check_dependencies() {
    if ! command -v ros &> /dev/null; then
        log_error "RouterOS CLI tool 'ros' not found"
        log_info "You can install it with: go get -u gopkg.in/routeros.v2/cmd/ros"
        exit 1
    fi
}

# Function to simulate a PPPoE connection
simulate_connect() {
    local username=$1
    local password=$2
    
    log_info "Simulating PPPoE connection for user: $username"
    
    # Note: In a real scenario, you would use a PPPoE client to connect
    # For this simulation, we're just logging the event
    # You could use pppd or similar tools to create actual connections
    
    # Example with ros CLI (if available):
    # ros -user=$CHR_USERNAME -pass=$CHR_PASSWORD -host=$CHR_HOST:$CHR_API_PORT \
    #     /ppp/active/print
    
    log_info "User $username would connect here (simulation)"
    sleep 1
}

# Function to simulate a PPPoE disconnection
simulate_disconnect() {
    local username=$1
    
    log_info "Simulating PPPoE disconnection for user: $username"
    
    # In real scenario, disconnect the PPPoE session
    log_info "User $username would disconnect here (simulation)"
    sleep 1
}

# Function to simulate random activity
simulate_random_activity() {
    local duration=$1
    local interval=$2
    
    log_info "Starting random PPPoE activity simulation for $duration seconds"
    log_info "Activity interval: $interval seconds"
    
    local end_time=$(($(date +%s) + duration))
    local active_users=()
    
    while [ $(date +%s) -lt $end_time ]; do
        # Randomly decide to connect or disconnect a user
        local action=$((RANDOM % 3))
        
        if [ $action -eq 0 ] && [ ${#active_users[@]} -lt ${#USERS[@]} ]; then
            # Connect a random user that's not already connected
            local available_users=()
            for user in "${USERS[@]}"; do
                if [[ ! " ${active_users[@]} " =~ " ${user} " ]]; then
                    available_users+=("$user")
                fi
            done
            
            if [ ${#available_users[@]} -gt 0 ]; then
                local random_index=$((RANDOM % ${#available_users[@]}))
                local user_to_connect="${available_users[$random_index]}"
                
                # Find password for this user
                for i in "${!USERS[@]}"; do
                    if [ "${USERS[$i]}" = "$user_to_connect" ]; then
                        simulate_connect "$user_to_connect" "${PASSWORDS[$i]}"
                        active_users+=("$user_to_connect")
                        break
                    fi
                done
            fi
        elif [ $action -eq 1 ] && [ ${#active_users[@]} -gt 0 ]; then
            # Disconnect a random active user
            local random_index=$((RANDOM % ${#active_users[@]}))
            local user_to_disconnect="${active_users[$random_index]}"
            
            simulate_disconnect "$user_to_disconnect"
            
            # Remove user from active list using proper array manipulation
            local temp_array=()
            for user in "${active_users[@]}"; do
                if [ "$user" != "$user_to_disconnect" ]; then
                    temp_array+=("$user")
                fi
            done
            active_users=("${temp_array[@]}")
        fi
        
        # Log current state
        log_info "Active sessions: ${#active_users[@]} / ${#USERS[@]}"
        
        # Wait before next activity
        sleep $interval
    done
    
    # Disconnect all remaining users
    log_info "Simulation ending, disconnecting all active users..."
    for user in "${active_users[@]}"; do
        simulate_disconnect "$user"
    done
    
    log_info "Simulation complete!"
}

# Function to simulate steady-state load
simulate_steady_load() {
    local num_users=$1
    local duration=$2
    
    log_info "Simulating steady load with $num_users concurrent users for $duration seconds"
    
    # Connect users
    for i in $(seq 0 $((num_users - 1))); do
        simulate_connect "${USERS[$i]}" "${PASSWORDS[$i]}"
    done
    
    log_info "Maintaining $num_users active sessions..."
    sleep $duration
    
    # Disconnect users
    log_info "Disconnecting all users..."
    for i in $(seq 0 $((num_users - 1))); do
        simulate_disconnect "${USERS[$i]}"
    done
    
    log_info "Steady load simulation complete!"
}

# Function to simulate peak hour traffic
simulate_peak_hour() {
    log_info "Simulating peak hour traffic pattern..."
    
    # Ramp up phase (connect users gradually)
    log_info "Ramp up phase: connecting users..."
    for i in "${!USERS[@]}"; do
        simulate_connect "${USERS[$i]}" "${PASSWORDS[$i]}"
        sleep $((RANDOM % 5 + 1))
    done
    
    # Peak phase (all users connected)
    log_info "Peak phase: maintaining full load..."
    sleep 30
    
    # Ramp down phase (disconnect users gradually)
    log_info "Ramp down phase: disconnecting users..."
    for i in "${!USERS[@]}"; do
        simulate_disconnect "${USERS[$i]}"
        sleep $((RANDOM % 5 + 1))
    done
    
    log_info "Peak hour simulation complete!"
}

# Function to query active sessions via RouterOS API
query_active_sessions() {
    log_info "Querying active PPPoE sessions..."
    
    if command -v ros &> /dev/null; then
        ros -user=$CHR_USERNAME -pass=$CHR_PASSWORD -host=$CHR_HOST:$CHR_API_PORT \
            /ppp/active/print 2>/dev/null || log_warn "Failed to query sessions"
    else
        log_warn "RouterOS CLI tool not available, skipping session query"
    fi
}

# Main menu
show_menu() {
    echo ""
    echo "=================================="
    echo "PPPoE Activity Simulation"
    echo "=================================="
    echo "1. Random activity (continuous)"
    echo "2. Steady load test"
    echo "3. Peak hour simulation"
    echo "4. Query active sessions"
    echo "5. Exit"
    echo "=================================="
}

# Main script
main() {
    log_info "PPPoE Activity Simulator for MikroTik CHR"
    log_info "Target: $CHR_HOST:$CHR_API_PORT"
    
    # Check dependencies
    # check_dependencies
    
    while true; do
        show_menu
        read -p "Select option: " choice
        
        case $choice in
            1)
                read -p "Duration (seconds, default 300): " duration
                duration=${duration:-300}
                read -p "Activity interval (seconds, default 10): " interval
                interval=${interval:-10}
                simulate_random_activity $duration $interval
                ;;
            2)
                read -p "Number of concurrent users (max ${#USERS[@]}, default 5): " num_users
                num_users=${num_users:-5}
                read -p "Duration (seconds, default 60): " duration
                duration=${duration:-60}
                simulate_steady_load $num_users $duration
                ;;
            3)
                simulate_peak_hour
                ;;
            4)
                query_active_sessions
                ;;
            5)
                log_info "Exiting..."
                exit 0
                ;;
            *)
                log_error "Invalid option"
                ;;
        esac
    done
}

# Run main function
main
