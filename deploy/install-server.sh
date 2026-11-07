#!/bin/sh
#
# Installs or upgrades pocketbaseserver and the SkillSurvey frontend on an OpenBSD app host. Run it as root from
# an unpacked server bundle, as often as you like: a second run with the same bundle and options changes nothing.
# deploy/README.md in the SkillSurvey repository gives the commands, the steps left by hand and how to go back.

set -eu
umask 022

script_name=install-server.sh
bundle_directory=$(cd "$(dirname "$0")" && pwd)
. "$bundle_directory/install-common.sh"

service_account=_skillsurvey
service_account_identifier=498
service_login_class=daemon
service_home_path=/var/skillsurvey
pocketbase_port=8090
data_folder_path=$service_home_path/pb_data
snapshots_folder_path=$service_home_path/snapshots
snapshots_to_keep=3
program_path=/usr/local/libexec/skillsurvey/pocketbaseserver
web_root_inside_httpd_chroot=/htdocs/skillsurvey
httpd_configuration_marker="# Written by install-server.sh, which replaces this file."
health_check_attempts=30
seconds_between_health_checks=2

service_home=$destination_root$service_home_path
data_folder=$destination_root$data_folder_path
snapshots_folder=$destination_root$snapshots_folder_path
installed_program=$destination_root$program_path
service_definition=$destination_root/etc/rc.d/pocketbaseserver
addresses_file=$destination_root/etc/skillsurvey/server.conf
install_record=$destination_root/var/db/skillsurvey/BUNDLE-INFO
httpd_configuration=$destination_root/etc/httpd.conf
web_root=$destination_root/var/www$web_root_inside_httpd_chroot
releases_folder=$web_root/releases

ipv4_address_pattern='([0-9]{1,3}\.){3}[0-9]{1,3}'
web_address_pattern='https?://[A-Za-z0-9.-]+(:[0-9]{1,5})?'

listen_address=
site_address=
pocketbase_address=
snapshot_name=
candidate_httpd_configuration_file=
usage="usage: install-server.sh [--listen-address IP] [--site-address URL] [--pocketbase-address URL]"

parse_arguments() {
	while [ $# -gt 0 ]; do
		case "$1" in
		--listen-address | --site-address | --pocketbase-address)
			[ $# -ge 2 ] || refuse "$1 needs a value. $usage"
			case "$1" in
			--listen-address) listen_address=$2 ;;
			--site-address) site_address=$2 ;;
			--pocketbase-address) pocketbase_address=$2 ;;
			esac
			shift 2
			;;
		*)
			refuse "$usage"
			;;
		esac
	done
}

read_saved_address() {
	[ -f "$addresses_file" ] || return 0
	sed -n "s/^$1=//p" "$addresses_file"
}

use_saved_addresses_for_any_not_given() {
	[ -n "$listen_address" ] || listen_address=$(read_saved_address listen_address)
	[ -n "$site_address" ] || site_address=$(read_saved_address site_address)
	[ -n "$pocketbase_address" ] || pocketbase_address=$(read_saved_address pocketbase_address)
	missing_options=
	[ -n "$listen_address" ] || missing_options="$missing_options --listen-address"
	[ -n "$site_address" ] || missing_options="$missing_options --site-address"
	[ -n "$pocketbase_address" ] || missing_options="$missing_options --pocketbase-address"
	[ -z "$missing_options" ] || refuse "a first install needs all three addresses, and these are missing:$missing_options"
	refuse_unless_value_matches --listen-address "$listen_address" "$ipv4_address_pattern"
	refuse_unless_value_matches --site-address "$site_address" "$web_address_pattern"
	refuse_unless_value_matches --pocketbase-address "$pocketbase_address" "$web_address_pattern"
}

address_the_checks_use() {
	if [ "$listen_address" = 0.0.0.0 ]; then
		echo 127.0.0.1
	else
		echo "$listen_address"
	fi
}

desired_pocketbase_flags() {
	echo "serve --http=$listen_address:$pocketbase_port --origins=$site_address --dir=$data_folder_path"
}

desired_addresses_file() {
	printf 'listen_address=%s\nsite_address=%s\npocketbase_address=%s' "$listen_address" "$site_address" "$pocketbase_address"
}

desired_host_configuration() {
	printf '{ "PocketBaseUrl": "%s" }' "$pocketbase_address"
}

desired_httpd_configuration() {
	httpd_listen_address=$listen_address
	[ "$listen_address" != 0.0.0.0 ] || httpd_listen_address='*'
	cat <<HTTPD_CONFIGURATION
$httpd_configuration_marker
server "default" {
	listen on $httpd_listen_address port 80
	root "$web_root_inside_httpd_chroot/current"
	location not found "/*" {
		request rewrite "/index.html"
	}
}
HTTPD_CONFIGURATION
}

installed_commit_short() {
	if [ -f "$install_record" ]; then
		sed -n 's/^commit=//p' "$install_record" | cut -c1-7
	else
		echo none
	fi
}

program_differs() {
	! { [ -f "$installed_program" ] && cmp -s "$bundle_directory/pocketbaseserver" "$installed_program"; }
}

data_folder_holds_a_database() {
	[ -f "$data_folder/data.db" ]
}

refuse_unless_httpd_configuration_is_ours_and_valid() {
	if [ -f "$httpd_configuration" ] && [ "$(head -1 "$httpd_configuration")" != "$httpd_configuration_marker" ]; then
		refuse "${httpd_configuration#"$destination_root"} exists and this script did not write it, so it was left alone and nothing was changed"
	fi
	if [ -f "$httpd_configuration" ] && [ "$(cat "$httpd_configuration")" = "$(desired_httpd_configuration)" ]; then
		return
	fi
	candidate_httpd_configuration_file=$(mktemp -t httpd.conf.XXXXXX)
	partial_file=$candidate_httpd_configuration_file
	desired_httpd_configuration >"$candidate_httpd_configuration_file"
	httpd -n -f "$candidate_httpd_configuration_file" >/dev/null 2>&1 ||
		refuse "httpd -n rejected the new httpd.conf, so nothing was changed: $(cat "$candidate_httpd_configuration_file")"
	rm -f "$candidate_httpd_configuration_file"
	partial_file=
}

refuse_unless_there_is_room_for_a_snapshot() {
	program_differs && data_folder_holds_a_database || return 0
	data_kilobytes=$(du -sk "$data_folder" | awk '{ print $1 }')
	needed_kilobytes=$((data_kilobytes + data_kilobytes / 10 + 1))
	available_kilobytes=$(df -kP "$service_home" | awk 'NR == 2 { print $4 }')
	[ "$available_kilobytes" -ge "$needed_kilobytes" ] ||
		refuse "a snapshot of pb_data needs ${needed_kilobytes} KB free on /var and only ${available_kilobytes} KB is free, so nothing was stopped or changed"
}

create_folders_and_service_definition() {
	begin_step "creating the account, folders and service definition"
	create_account_if_missing "$service_account" "$service_account_identifier" "$service_login_class" "$service_home_path" "SkillSurvey PocketBase"
	ensure_directory "$service_home" 0755 root:wheel
	if [ ! -d "$data_folder" ]; then
		mkdir "$data_folder"
		chmod 0700 "$data_folder"
		chown "$service_account:$service_account" "$data_folder"
		report_change "created $data_folder_path"
	fi
	ensure_directory "$snapshots_folder" 0700 root:wheel
	ensure_directory "$(dirname "$installed_program")" 0755 root:wheel
	ensure_directory "$(dirname "$install_record")" 0755 root:wheel
	ensure_directory "$(dirname "$addresses_file")" 0755 root:wheel
	install_file_if_it_differs "$bundle_directory/pocketbaseserver.rc" "$service_definition" 0555 root:wheel
	install_content_if_it_differs "$(desired_addresses_file)" "$addresses_file" 0644 root:wheel
}

wait_for_pocketbase_to_answer() {
	health_address="http://$(address_the_checks_use):$pocketbase_port/api/health"
	attempt=1
	while [ "$attempt" -le "$health_check_attempts" ]; do
		if ftp -V -w 5 -o /dev/null "$health_address" >/dev/null 2>&1; then
			return 0
		fi
		sleep "$seconds_between_health_checks"
		attempt=$((attempt + 1))
	done
	return 1
}

explain_failed_health_check_and_exit() {
	failure_already_explained=yes
	{
		echo "$script_name: pocketbaseserver did not answer $health_address within $((health_check_attempts * seconds_between_health_checks)) seconds."
		echo "Nothing was put back. Its log is in /var/log/daemon."
		if [ -n "$snapshot_name" ]; then
			echo "The data from before this upgrade is in $snapshots_folder_path/$snapshot_name. To go back:"
			echo "  doas rcctl stop pocketbaseserver"
			echo "  doas mv $data_folder_path $data_folder_path.failed"
			echo "  doas cp -Rp $snapshots_folder_path/$snapshot_name/pb_data $data_folder_path"
			echo "  doas sh <the previous bundle>/install-server.sh"
		else
			echo "No snapshot was taken in this run, because the program did not change or there was no database."
		fi
	} >&2
	exit 1
}

copy_data_to_a_snapshot_or_start_the_installed_program_again() {
	begin_step "copying pb_data to a snapshot"
	snapshot_name="$(date -u +%Y%m%dT%H%M%SZ)-$(installed_commit_short)"
	partial_directory="$snapshots_folder/.partial-$snapshot_name"
	mkdir "$partial_directory"
	if ! cp -Rp "$data_folder" "$partial_directory/pb_data" 2>/dev/null; then
		rm -rf "$partial_directory"
		partial_directory=
		snapshot_name=
		failure_already_explained=yes
		echo "$script_name: could not copy $data_folder_path to a snapshot, so nothing was replaced" >&2
		if [ -f "$installed_program" ]; then
			rcctl start pocketbaseserver >/dev/null || true
			echo "$script_name: started the installed pocketbaseserver again" >&2
		fi
		exit 1
	fi
	mv "$partial_directory" "$snapshots_folder/$snapshot_name"
	partial_directory=
	report_change "copied $data_folder_path to $snapshots_folder_path/$snapshot_name"
}

install_and_start_pocketbase() {
	begin_step "installing pocketbaseserver"
	enable_service_if_disabled pocketbaseserver
	current_flags=$(rcctl get pocketbaseserver flags)
	flags_differ=no
	[ "$current_flags" = "$(desired_pocketbase_flags)" ] || flags_differ=yes
	program_will_be_replaced=no
	! program_differs || program_will_be_replaced=yes

	if [ "$program_will_be_replaced" = no ] && [ "$flags_differ" = no ]; then
		if service_is_running pocketbaseserver; then
			return
		fi
		begin_step "starting pocketbaseserver"
		rcctl start pocketbaseserver >/dev/null
		report_change "started pocketbaseserver"
		wait_for_pocketbase_to_answer || explain_failed_health_check_and_exit
		return
	fi

	# rc.d finds a running service by its exact command line, so new flags set first would make the stop miss it.
	if service_is_running pocketbaseserver; then
		begin_step "stopping pocketbaseserver"
		rcctl stop pocketbaseserver >/dev/null ||
			refuse "rcctl stop pocketbaseserver failed, so the program and data were left alone"
		report_change "stopped pocketbaseserver"
	fi
	if [ "$program_will_be_replaced" = yes ] && data_folder_holds_a_database; then
		copy_data_to_a_snapshot_or_start_the_installed_program_again
	fi
	if [ "$program_will_be_replaced" = yes ]; then
		begin_step "replacing pocketbaseserver"
		install_file_if_it_differs "$bundle_directory/pocketbaseserver" "$installed_program" 0755 root:wheel
	fi
	if [ "$flags_differ" = yes ]; then
		begin_step "setting pocketbaseserver's flags"
		# shellcheck disable=SC2046 # the flags are separate words, and were checked to hold no spaces or quotes
		rcctl set pocketbaseserver flags $(desired_pocketbase_flags)
		report_change "set pocketbaseserver flags to $(desired_pocketbase_flags)"
	fi
	begin_step "starting pocketbaseserver"
	rcctl start pocketbaseserver >/dev/null
	report_change "started pocketbaseserver"
	wait_for_pocketbase_to_answer || explain_failed_health_check_and_exit
}

live_release() {
	readlink "$web_root/current" 2>/dev/null || true
}

live_release_matches_the_bundle() {
	release=$(live_release)
	[ -n "$release" ] && [ -d "$web_root/$release" ] &&
		diff -r -x host-config.json "$bundle_directory/frontend" "$web_root/$release" >/dev/null 2>&1
}

unpack_a_new_release_and_switch_to_it() {
	release_name=$(bundle_commit | cut -c1-7)
	previous_release=$(live_release)
	if [ "releases/$release_name" = "$previous_release" ]; then
		release_name="$release_name.$(date -u +%Y%m%dT%H%M%SZ)"
	fi
	partial_directory=$(mktemp -d "$releases_folder/.partial.XXXXXX")
	cp -R "$bundle_directory/frontend/." "$partial_directory/"
	desired_host_configuration >"$partial_directory/host-config.json"
	echo >>"$partial_directory/host-config.json"
	find "$partial_directory" -type d -exec chmod 0755 {} +
	find "$partial_directory" -type f -exec chmod 0644 {} +
	chown -R root:daemon "$partial_directory"
	rm -rf "$releases_folder/$release_name"
	mv "$partial_directory" "$releases_folder/$release_name"
	partial_directory=
	if [ -n "$previous_release" ]; then
		ln -shf "$previous_release" "$web_root/previous"
	fi
	ln -shf "releases/$release_name" "$web_root/current"
	report_change "switched the frontend to releases/$release_name"
}

install_frontend() {
	begin_step "installing the frontend"
	ensure_directory "$web_root" 0755 root:daemon
	ensure_directory "$releases_folder" 0755 root:daemon
	if live_release_matches_the_bundle; then
		install_content_if_it_differs "$(desired_host_configuration)" "$web_root/$(live_release)/host-config.json" 0644 root:daemon
	else
		unpack_a_new_release_and_switch_to_it
	fi
}

configure_and_start_httpd() {
	begin_step "configuring httpd"
	httpd_configuration_changed=no
	if ! { [ -f "$httpd_configuration" ] && [ "$(cat "$httpd_configuration")" = "$(desired_httpd_configuration)" ]; }; then
		desired_httpd_configuration | install_standard_input_atomically "$httpd_configuration" 0644 root:wheel
		report_change "wrote /etc/httpd.conf"
		httpd_configuration_changed=yes
	fi
	enable_service_if_disabled httpd
	if ! service_is_running httpd; then
		rcctl start httpd >/dev/null
		report_change "started httpd"
	elif [ "$httpd_configuration_changed" = yes ]; then
		rcctl reload httpd >/dev/null
		report_change "reloaded httpd"
	fi
}

refuse_unless_httpd_serves() {
	path=$1
	expected_file=$2
	address="http://$(address_the_checks_use)$path"
	answer_file=$(mktemp -t httpd-check.XXXXXX)
	partial_file=$answer_file
	if ! ftp -V -w 5 -o "$answer_file" "$address" >/dev/null 2>&1 || ! cmp -s "$answer_file" "$expected_file"; then
		refuse "httpd did not serve $address as expected. PocketBase is still running"
	fi
	rm -f "$answer_file"
	partial_file=
}

check_the_site_through_httpd() {
	begin_step "checking the site through httpd"
	live_release_folder=$web_root/$(live_release)
	refuse_unless_httpd_serves / "$live_release_folder/index.html"
	refuse_unless_httpd_serves /user/monthly-count-report "$live_release_folder/index.html"
	refuse_unless_httpd_serves /host-config.json "$live_release_folder/host-config.json"
}

remove_releases_other_than_live_and_previous() {
	begin_step "removing old releases"
	live_release_name=$(basename "$(live_release)")
	previous_release_name=$(basename "$(readlink "$web_root/previous" 2>/dev/null || echo none)")
	for release_folder in "$releases_folder"/* "$releases_folder"/.partial.*; do
		[ -d "$release_folder" ] || continue
		release_name=$(basename "$release_folder")
		[ "$release_name" != "$live_release_name" ] && [ "$release_name" != "$previous_release_name" ] || continue
		rm -rf "$release_folder"
		report_change "removed old release releases/$release_name"
	done
}

remove_snapshots_beyond_the_newest_kept() {
	[ -n "$snapshot_name" ] || return 0
	begin_step "removing old snapshots"
	# Only names this script makes are counted or removed, so a snapshot someone kept by hand under another name stays.
	old_snapshot_names=$(ls "$snapshots_folder" | grep -E '^[0-9]{8}T[0-9]{6}Z-[0-9a-z]+$' | sort -r | tail -n +$((snapshots_to_keep + 1)))
	for old_snapshot_name in $old_snapshot_names; do
		rm -rf "${snapshots_folder:?}/$old_snapshot_name"
		report_change "removed old snapshot $snapshots_folder_path/$old_snapshot_name"
	done
}

record_the_installed_bundle() {
	begin_step "recording the installed bundle"
	install_file_if_it_differs "$bundle_directory/BUNDLE-INFO" "$install_record" 0644 root:wheel
}

print_steps_left_by_hand() {
	pocketbase_dashboard="$pocketbase_address/_/"
	cat <<STEPS_LEFT_BY_HAND

This was a first install. These steps are left to do by hand:
  1. Create the first superuser. Its password shows in ps for a moment:
     doas -u $service_account $program_path superuser upsert EMAIL PASSWORD --dir=$data_folder_path
  2. In the dashboard at $pocketbase_dashboard, set a password on runtask@skillsurvey.com,
     for the runtask host's runtask.json.
  3. Open $site_address and $pocketbase_dashboard from another machine. Only that shows whether pf lets them through.
STEPS_LEFT_BY_HAND
}

refuse_unless_run_as_root
parse_arguments "$@"
use_saved_addresses_for_any_not_given
refuse_unless_bundle_was_built_for_this_host
refuse_unless_bundle_files_match_their_checksums
refuse_unless_existing_account_is_as_expected "$service_account" "$service_login_class" "$service_home_path" /sbin/nologin
refuse_unless_httpd_configuration_is_ours_and_valid
refuse_unless_there_is_room_for_a_snapshot

first_install=no
[ -f "$install_record" ] || first_install=yes

create_folders_and_service_definition
install_and_start_pocketbase
install_frontend
configure_and_start_httpd
check_the_site_through_httpd
remove_releases_other_than_live_and_previous
remove_snapshots_beyond_the_newest_kept
record_the_installed_bundle
[ "$first_install" = no ] || print_steps_left_by_hand
report_change_count
