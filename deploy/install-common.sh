# Functions shared by install-server.sh and install-runtask.sh. Each sets script_name and bundle_directory, then
# reads this file with ".", so it is never run on its own.

# DESTDIR exists for the tests, which run the scripts against a throwaway root.
destination_root=${DESTDIR:-}

changes=0
current_step="checking the bundle and this host"
failure_already_explained=no
partial_file=
partial_directory=
newline='
'

refuse() {
	failure_already_explained=yes
	echo "$script_name: $1" >&2
	exit 1
}

report_change() {
	changes=$((changes + 1))
	echo "$1"
}

report_change_count() {
	if [ "$changes" -eq 1 ]; then
		echo "1 change"
	else
		echo "$changes changes"
	fi
}

begin_step() {
	current_step=$1
}

remove_partial_work_and_name_the_failed_step() {
	exit_status=$?
	[ -z "$partial_file" ] || rm -f "$partial_file"
	[ -z "$partial_directory" ] || rm -rf "$partial_directory"
	if [ "$exit_status" -ne 0 ] && [ "$failure_already_explained" = no ]; then
		echo "$script_name: failed while $current_step" >&2
	fi
}
trap remove_partial_work_and_name_the_failed_step EXIT

refuse_unless_value_matches() {
	option_name=$1
	value=$2
	pattern=$3
	case $value in
	*"$newline"*) refuse "$option_name must be a single line" ;;
	esac
	printf '%s\n' "$value" | grep -Eqx "$pattern" || refuse "$option_name has an unexpected value: $value"
}

refuse_unless_run_as_root() {
	[ "$(id -u)" -eq 0 ] || refuse "run this as root"
}

read_bundle_information() {
	sed -n "s/^$1=//p" "$bundle_directory/BUNDLE-INFO"
}

refuse_unless_bundle_was_built_for_this_host() {
	[ -f "$bundle_directory/BUNDLE-INFO" ] || refuse "$bundle_directory/BUNDLE-INFO is missing, so this is not a whole bundle"
	host_system=$(uname -s)
	host_release=$(uname -r)
	host_architecture=$(uname -m)
	bundle_release=$(read_bundle_information built_on_release)
	bundle_architecture=$(read_bundle_information built_on_architecture)
	[ "$host_system" = OpenBSD ] || refuse "this host runs $host_system, and the bundle is for OpenBSD"
	[ "$host_release" = "$bundle_release" ] ||
		refuse "this host runs OpenBSD $host_release, and the bundle was built on OpenBSD $bundle_release"
	[ "$host_architecture" = "$bundle_architecture" ] ||
		refuse "this host is $host_architecture, and the bundle was built for $bundle_architecture"
}

refuse_unless_bundle_files_match_their_checksums() {
	[ -f "$bundle_directory/SHA256" ] || refuse "$bundle_directory/SHA256 is missing, so the bundle cannot be checked"
	if ! failed_checks=$(cd "$bundle_directory" && sha256 -q -c SHA256 2>&1); then
		refuse "these files do not match SHA256, so nothing was changed: $failed_checks"
	fi
}

bundle_commit() {
	read_bundle_information commit
}

refuse_unless_existing_account_is_as_expected() {
	account_name=$1
	expected_login_class=$2
	expected_home_directory=$3
	expected_shell=$4
	account=$(awk -F: -v user="$account_name" '$1 == user { print $5 ":" $9 ":" $10 }' "$destination_root/etc/master.passwd")
	[ -n "$account" ] || return 0

	account_login_class=${account%%:*}
	account_home_and_shell=${account#*:}
	account_home_directory=${account_home_and_shell%%:*}
	account_shell=${account_home_and_shell#*:}
	[ "$account_login_class" = "$expected_login_class" ] ||
		refuse "$account_name already exists with login class $account_login_class, not $expected_login_class"
	[ "$account_home_directory" = "$expected_home_directory" ] ||
		refuse "$account_name already exists with home $account_home_directory, not $expected_home_directory"
	[ "$account_shell" = "$expected_shell" ] ||
		refuse "$account_name already exists with shell $account_shell, not $expected_shell"
	if awk -F: '$1 == "wheel" { print $4 }' "$destination_root/etc/group" | tr ',' '\n' | grep -qx "$account_name"; then
		refuse "$account_name is in wheel, and a service account must not be, so remove it first"
	fi
}

create_account_if_missing() {
	account_name=$1
	account_identifier=$2
	account_login_class=$3
	account_home_directory=$4
	account_full_name=$5
	if grep -q "^$account_name:" "$destination_root/etc/master.passwd"; then
		return
	fi
	useradd -u "$account_identifier" -g =uid -L "$account_login_class" -d "$account_home_directory" \
		-s /sbin/nologin -c "$account_full_name" "$account_name"
	report_change "created account $account_name"
}

# The rename means a program already running keeps its old file, and nothing can see a half-written one.
install_standard_input_atomically() {
	target=$1
	mode=$2
	owner=$3
	partial_file=$(mktemp "$(dirname "$target")/.partial.XXXXXX")
	cat >"$partial_file"
	chmod "$mode" "$partial_file"
	chown "$owner" "$partial_file"
	mv -f "$partial_file" "$target"
	partial_file=
}

install_file_if_it_differs() {
	source_file=$1
	target=$2
	mode=$3
	owner=$4
	if [ -f "$target" ] && cmp -s "$source_file" "$target"; then
		return
	fi
	install_standard_input_atomically "$target" "$mode" "$owner" <"$source_file"
	report_change "installed ${target#"$destination_root"}"
}

install_content_if_it_differs() {
	content=$1
	target=$2
	mode=$3
	owner=$4
	if [ -f "$target" ] && [ "$(cat "$target")" = "$content" ]; then
		return
	fi
	printf '%s\n' "$content" | install_standard_input_atomically "$target" "$mode" "$owner"
	report_change "wrote ${target#"$destination_root"}"
}

ensure_directory() {
	path=$1
	mode=$2
	owner=$3
	if [ -d "$path" ]; then
		return
	fi
	mkdir -p "$path"
	chmod "$mode" "$path"
	chown "$owner" "$path"
	report_change "created ${path#"$destination_root"}"
}

enable_service_if_disabled() {
	service_name=$1
	if rcctl get "$service_name" status >/dev/null 2>&1; then
		return
	fi
	rcctl enable "$service_name"
	report_change "enabled $service_name at boot"
}

service_is_running() {
	rcctl check "$1" >/dev/null 2>&1
}
