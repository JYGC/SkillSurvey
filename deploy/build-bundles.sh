#!/bin/sh
#
# Builds the server and runtask bundles for whatever commit this checkout is on, from an export of that commit, so
# uncommitted files cannot enter a bundle. In CI the bundles go to $CI_RUN_ARTIFACT_DIRECTORY. By hand they go to
# a skillsurvey-bundles folder beside the checkout, outside it so a bundle can never be committed.
# A commit whose bundles already exist is not built again, so one commit always means the same files.

set -eu
umask 022

checkout=$(cd "$(dirname "$0")/.." && pwd)
output_folder=${CI_RUN_ARTIFACT_DIRECTORY:-$(dirname "$checkout")/skillsurvey-bundles}

export_folder=
partial_bundle_files=

refuse() {
	echo "build-bundles.sh: $1" >&2
	exit 1
}

remove_export_and_partial_bundles() {
	[ -z "$export_folder" ] || rm -rf "$export_folder"
	for partial_bundle_file in $partial_bundle_files; do
		rm -f "$partial_bundle_file"
	done
}
trap remove_export_and_partial_bundles EXIT

commit=$(git -C "$checkout" rev-parse HEAD)
short_commit=$(echo "$commit" | cut -c1-7)
server_bundle_name=skillsurvey-server-$short_commit
runtask_bundle_name=skillsurvey-runtask-$short_commit

if [ -f "$output_folder/$server_bundle_name.tgz" ] && [ -f "$output_folder/$runtask_bundle_name.tgz" ]; then
	echo "both bundles for $short_commit already exist in $output_folder, so they were kept"
	exit 0
fi
git -C "$checkout" cat-file -e "$commit:deploy" 2>/dev/null ||
	refuse "commit $short_commit has no deploy folder, so it cannot be bundled"

if [ -n "$(git -C "$checkout" status --porcelain)" ]; then
	echo "the checkout has uncommitted changes, which are left out: only commit $short_commit is built"
fi

mkdir -p "$output_folder"
export_folder=$(mktemp -d -t skillsurvey-bundle-export.XXXXXX)
git -C "$checkout" archive "$commit" | tar -xf - -C "$export_folder"
staging_folder=$export_folder/.bundles
mkdir "$staging_folder"

# Node needs more than the 1.5 GB default data limit to build the frontend.
ulimit -d 4194304
ulimit -c 0

echo "building the frontend for $short_commit"
(cd "$export_folder/frontend" && npm ci --no-audit --no-fund && npm run build:hosted)
development_address=$(sed -n 's/^VUE_APP_POCKETBASE_URL=//p' "$export_folder/frontend/.env" 2>/dev/null || true)
if [ -n "$development_address" ] && grep -rqF "$development_address" "$export_folder/frontend/dist"; then
	refuse "the built frontend holds the development address from frontend/.env, so no bundle was made"
fi

# -trimpath and -buildvcs=false make the programs byte-identical wherever they are built. Without them every commit
# would look like a new pocketbaseserver to the install script, which would stop it and take a snapshot for nothing.
echo "building the programs for $short_commit"
mkdir "$staging_folder/$server_bundle_name" "$staging_folder/$runtask_bundle_name"
(cd "$export_folder/pocketbaseserver" && go build -trimpath -buildvcs=false -o "$staging_folder/$server_bundle_name/pocketbaseserver" ./cmd/pocketbaseserver)
(cd "$export_folder/runtask" && go build -trimpath -buildvcs=false -o "$staging_folder/$runtask_bundle_name/runtask" ./cmd/runtask)

write_bundle_information() {
	printf 'commit=%s\nbuilt_on_release=%s\nbuilt_on_architecture=%s\n' "$commit" "$(uname -r)" "$(uname -m)" >"$1/BUNDLE-INFO"
}

write_checksums_of_every_other_file() {
	(cd "$1" && find . -type f ! -name SHA256 | sed 's|^\./||' | sort | xargs sha256 >SHA256)
}

server_bundle=$staging_folder/$server_bundle_name
cp "$export_folder/deploy/install-server.sh" "$export_folder/deploy/install-common.sh" "$export_folder/deploy/pocketbaseserver.rc" "$server_bundle/"
cp -R "$export_folder/frontend/dist" "$server_bundle/frontend"
write_bundle_information "$server_bundle"

runtask_bundle=$staging_folder/$runtask_bundle_name
cp "$export_folder/deploy/install-runtask.sh" "$export_folder/deploy/install-common.sh" "$export_folder/deploy/run-scheduled-job" "$runtask_bundle/"
cp -R "$export_folder/deploy/examples" "$runtask_bundle/examples"
write_bundle_information "$runtask_bundle"

for bundle in "$server_bundle" "$runtask_bundle"; do
	find "$bundle" -type d -exec chmod 0755 {} +
	find "$bundle" -type f -exec chmod 0644 {} +
	chmod 0755 "$bundle"/pocketbaseserver "$bundle"/runtask "$bundle"/run-scheduled-job 2>/dev/null || true
	write_checksums_of_every_other_file "$bundle"
done

for bundle_name in "$server_bundle_name" "$runtask_bundle_name"; do
	partial_bundle_file=$output_folder/.partial-$bundle_name.tgz
	partial_bundle_files="$partial_bundle_files $partial_bundle_file"
	tar -czf "$partial_bundle_file" -C "$staging_folder" "$bundle_name"
done

# Both are renamed into place only after both were packed, so a failure never leaves one bundle without the other.
mv "$output_folder/.partial-$server_bundle_name.tgz" "$output_folder/$server_bundle_name.tgz"
mv "$output_folder/.partial-$runtask_bundle_name.tgz" "$output_folder/$runtask_bundle_name.tgz"
partial_bundle_files=
echo "built $output_folder/$server_bundle_name.tgz"
echo "built $output_folder/$runtask_bundle_name.tgz"
