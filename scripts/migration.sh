# Wrapper around Goose

# check argument is valid
if [ "$1" != "up" ] && [ "$1" != down ]; then
	>&2 echo "bad argument, can only be 'up' or 'down'"
	exit 1
fi

# load env values
if ! source .env; then
	>&2 echo "failed to load .env"
	exit 1
fi

# go to where migrations are
pushd sql/schema || exit 1
goose postgres "$DB_URL" "$1"
popd || exit 1
