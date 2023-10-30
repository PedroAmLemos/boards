#!/bin/bash

# Function to run the command in a new split pane
run_in_tmux_pane() {
    tmux split-window -h "$1"
    tmux select-layout even-horizontal
}

# Check if we are inside a tmux session
if [ -z "$TMUX" ]; then
    # Outside of tmux, start a new session with a new window and the first command
    tmux new-session -d 'go run . node_1 hosts.local'
    
    # Switch to the new window and run the other commands in split panes
    tmux new-window
    tmux send-keys 'go run . node_1 hosts.local' C-m
    run_in_tmux_pane 'go run . node_2 hosts.local'
    run_in_tmux_pane 'go run . node_3 hosts.local'
    
    # Attach to the tmux session
    tmux attach
else
    # Inside a tmux session, create a new window and run the commands in split panes
    tmux new-window
    tmux send-keys 'go run . node_1 hosts.local' C-m
    run_in_tmux_pane 'go run . node_2 hosts.local'
    run_in_tmux_pane 'go run . node_3 hosts.local'
fi

