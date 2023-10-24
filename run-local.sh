#!/bin/bash
go build .

# Check if we are inside a tmux session
if [ -z "$TMUX" ]; then
    # Outside of tmux

    # Check if the tmux session exists
    tmux has-session -t multicast 2>/dev/null

    # $? is a special variable that captures the exit status of the last command
    if [ $? != 0 ]; then
      # If the session doesn't exist, create a new session. No need to attach yet.
      tmux new-session -d -s multicast
    fi

    # Create a new window in the session and run the first command
    tmux new-window -t multicast -n 'boards' './boards node_1 hosts.local'

    # Split the new window vertically and run the second command
    tmux split-window -t multicast:boards -h './boards node_2 hosts.local'

    # Split the pane again (from the first pane) and run the third command
    tmux split-window -t multicast:boards -h './boards node_3 hosts.local'

    # Set the layout to ensure all panes have equal width
    tmux select-layout -t multicast:boards even-horizontal

    # Attach to the new window (this will also attach to the session)
    tmux select-window -t multicast:boards
    tmux attach -t multicast
else
    # Inside of tmux

    # Create a new window and run the first command
    tmux new-window -n 'boards' './boards node_1 hosts.local'

    # Split the new window vertically and run the second command
    tmux split-window -h './boards node_2 hosts.local'

    # Split the pane again (from the first pane) and run the third command
    tmux split-window -h './boards node_3 hosts.local'

    # Set the layout to ensure all panes have equal width
    tmux select-layout even-horizontal

    # Switch to the new window
    tmux select-window -t boards
fi


