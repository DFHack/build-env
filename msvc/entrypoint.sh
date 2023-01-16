#!/bin/bash

set -e

# Default behaviour if BUILDER_UID is 0
if [[ -z "$BUILDER_UID" ]]; then
    exec "$@"
fi

# We can really only change users if we're root, so if we're not, just fallback
# to running the command
if [[ $(id -u) -ne 0 ]]; then
    exec "$@"
fi

OLD_UID=$(id -u buildmaster)
# Now we are root, and have BUILDER_UID set. If the UIDs mismatch for
# buildmaster, let's fix it.
if [[ $OLD_UID -ne $BUILDER_UID ]]; then
    echo "Changing builder UID from $OLD_UID to $BUILDER_UID"
    userdel buildmaster

    useradd -u $BUILDER_UID --shell /bin/bash buildmaster

    find / -xdev -uid $OLD_UID -exec chown buildmaster {} +
fi

# Preserve PATH cause su overrides it
exec su buildmaster -c 'env PATH='"$PATH"' "$0" "$@"' -- "$@"
