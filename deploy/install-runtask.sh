#!/bin/sh
#
# Installs or upgrades runtask, Chromium and the job schedule on an OpenBSD runtask host. Run it as root from an
# unpacked runtask bundle, as often as you like: a second run with the same bundle and options changes nothing.
# It never overwrites a settings file, and it changes the schedule only when given --start-schedule.
# deploy/README.md in the SkillSurvey repository gives the commands and the steps left by hand.

set -eu
umask 022

script_name=install-runtask.sh
bundle_directory=$(cd "$(dirname "$0")" && pwd)
. "$bundle_directory/install-common.sh"

service_account=_runtask
service_account_identifier=497
service_login_class=runtask
service_home_path=/var/runtask
login_class_definition="$service_login_class:datasize=4096M:coredumpsize=0:maxproc=256:tc=default:"
program_folder_path=/usr/local/libexec/runtask
job_runner_path=$program_folder_path/run-scheduled-job
settings_folder_path=/etc/runtask
settings_file_names="runtask.json www.seek.com.au.json au.jora.com.json"
x_libraries_chromium_needs="libX11 libxcb libXext libXrandr libfontconfig libfreetype libdrm libgbm"
schedule_begin_line="# BEGIN runtask schedule. install-runtask.sh replaces everything down to the END line."
schedule_end_line="# END runtask schedule"

service_home=$destination_root$service_home_path
log_folder=$destination_root/var/log/runtask
login_class_file=$destination_root/etc/login.conf.d/$service_login_class
program_folder=$destination_root$program_folder_path
settings_folder=$destination_root$settings_folder_path
install_record=$destination_root/var/db/runtask/BUNDLE-INFO
system_crontab=$destination_root/etc/crontab
x_library_folder=$destination_root/usr/X11R6/lib

scrape_hours_pattern='([0-9]|1[0-9]|2[0-3])(,([0-9]|1[0-9]|2[0-3]))*'

start_schedule=no
scrape_hours=

usage="usage: install-runtask.sh [--start-schedule --scrape-hours HOURS], such as --scrape-hours 7,17"

parse_arguments() {
	while [ $# -gt 0 ]; do
		case "$1" in
		--start-schedule)
			start_schedule=yes
			shift
			;;
		--scrape-hours)
			[ $# -ge 2 ] || refuse "--scrape-hours needs a value. $usage"
			scrape_hours=$2
			shift 2
			;;
		*)
			refuse "$usage"
			;;
		esac
	done
	if [ "$start_schedule" = yes ] && [ -z "$scrape_hours" ]; then
		refuse "--start-schedule needs --scrape-hours HOURS. $usage"
	fi
	if [ "$start_schedule" = no ] && [ -n "$scrape_hours" ]; then
		refuse "--scrape-hours only applies with --start-schedule. $usage"
	fi
	[ -z "$scrape_hours" ] || refuse_unless_value_matches --scrape-hours "$scrape_hours" "$scrape_hours_pattern"
}

refuse_unless_x_libraries_are_present() {
	missing_libraries=
	for library_name in $x_libraries_chromium_needs; do
		set -- "$x_library_folder/$library_name".so.*
		[ -e "$1" ] || missing_libraries="$missing_libraries $library_name"
	done
	[ -z "$missing_libraries" ] ||
		refuse "Chromium needs these X libraries, which are missing, so install the X sets first:$missing_libraries"
}

create_login_class_account_and_folders() {
	begin_step "creating the login class, account and folders"
	ensure_directory "$(dirname "$login_class_file")" 0755 root:wheel
	if [ ! -f "$login_class_file" ]; then
		printf '%s\n' "$login_class_definition" | install_standard_input_atomically "$login_class_file" 0644 root:wheel
		report_change "created ${login_class_file#"$destination_root"}"
	fi
	create_account_if_missing "$service_account" "$service_account_identifier" "$service_login_class" "$service_home_path" "SkillSurvey runtask"
	ensure_directory "$service_home" 0700 "$service_account:$service_account"
	ensure_directory "$log_folder" 0750 "$service_account:wheel"
	ensure_directory "$program_folder" 0755 root:wheel
	ensure_directory "$settings_folder" 0750 "root:$service_account"
	ensure_directory "$(dirname "$install_record")" 0755 root:wheel
}

install_chromium_if_missing() {
	begin_step "installing chromium"
	if pkg_info -e 'chromium-*' >/dev/null 2>&1; then
		return
	fi
	pkg_add chromium || refuse "pkg_add chromium failed, so no schedule was added or changed"
	report_change "installed the chromium package"
}

install_programs_and_settings() {
	begin_step "installing runtask"
	install_file_if_it_differs "$bundle_directory/runtask" "$program_folder/runtask" 0755 root:wheel
	install_file_if_it_differs "$bundle_directory/run-scheduled-job" "$program_folder/run-scheduled-job" 0755 root:wheel
	# runtask reads runtask.json only from beside its own program, and settings belong in /etc.
	if [ "$(readlink "$program_folder/runtask.json" 2>/dev/null || true)" != "$settings_folder_path/runtask.json" ]; then
		ln -sf "$settings_folder_path/runtask.json" "$program_folder/runtask.json"
		report_change "linked $program_folder_path/runtask.json to $settings_folder_path/runtask.json"
	fi

	begin_step "creating missing settings files from their examples"
	for settings_file_name in $settings_file_names; do
		[ ! -f "$settings_folder/$settings_file_name" ] || continue
		install_standard_input_atomically "$settings_folder/$settings_file_name" 0640 "root:$service_account" <"$bundle_directory/examples/$settings_file_name"
		report_change "created $settings_folder_path/$settings_file_name from its example"
	done
}

record_the_installed_bundle() {
	begin_step "recording the installed bundle"
	install_file_if_it_differs "$bundle_directory/BUNDLE-INFO" "$install_record" 0644 root:wheel
}

desired_schedule_block() {
	job="$service_account -n -s $job_runner_path"
	cat <<SCHEDULE
$schedule_begin_line
MAILTO=root
0 $scrape_hours * * * $job scrape
0 1 1 * * $job report
0 19 * * 5 $job housekeeping cleanfs
0 19 * * 5 $job housekeeping sendlog
$schedule_end_line
SCHEDULE
}

refuse_unless_settings_hold_no_placeholders() {
	settings_with_placeholders=
	for settings_file_name in $settings_file_names; do
		if grep -q REPLACE_ME "$settings_folder/$settings_file_name"; then
			settings_with_placeholders="$settings_with_placeholders $settings_folder_path/$settings_file_name"
		fi
	done
	[ -z "$settings_with_placeholders" ] ||
		refuse "these settings still hold REPLACE_ME, so the schedule was not changed:$settings_with_placeholders"
}

refuse_unless_check_setup_passes() {
	begin_step "running runtask check-setup as $service_account"
	su -c "$service_login_class" -s /bin/sh "$service_account" -c "$job_runner_path check-setup" ||
		refuse "runtask check-setup failed, so the schedule was not changed"
}

crontab_with_desired_schedule() {
	if [ ! -f "$system_crontab" ]; then
		desired_schedule_block
		return
	fi
	begin_line_count=$(grep -cxF "$schedule_begin_line" "$system_crontab" || true)
	end_line_count=$(grep -cxF "$schedule_end_line" "$system_crontab" || true)
	if [ "$begin_line_count" -eq 0 ] && [ "$end_line_count" -eq 0 ]; then
		cat "$system_crontab"
		[ -z "$(tail -c 1 "$system_crontab")" ] || echo
		desired_schedule_block
		return
	fi
	[ "$begin_line_count" -eq 1 ] && [ "$end_line_count" -eq 1 ] ||
		refuse "${system_crontab#"$destination_root"} has $begin_line_count BEGIN and $end_line_count END lines for the runtask schedule, so it was left alone"
	schedule_block=$(desired_schedule_block) begin_line=$schedule_begin_line end_line=$schedule_end_line awk '
		$0 == ENVIRON["begin_line"] { print ENVIRON["schedule_block"]; replacing = 1; next }
		replacing && $0 == ENVIRON["end_line"] { replacing = 0; next }
		replacing { next }
		{ print }
	' "$system_crontab"
}

start_schedule_if_asked() {
	if [ "$start_schedule" = no ]; then
		echo "schedule left as it was: pass --start-schedule --scrape-hours HOURS to set it"
		return
	fi
	begin_step "checking the settings before changing the schedule"
	refuse_unless_settings_hold_no_placeholders
	refuse_unless_check_setup_passes
	begin_step "writing the schedule"
	desired_crontab=$(crontab_with_desired_schedule)
	if [ -f "$system_crontab" ] && [ "$(cat "$system_crontab")" = "$desired_crontab" ]; then
		echo "schedule unchanged"
	else
		# cron ignores /etc/crontab if anyone but root can write to it, and notices a changed file by itself.
		printf '%s\n' "$desired_crontab" | install_standard_input_atomically "$system_crontab" 0644 root:wheel
		report_change "set the schedule in /etc/crontab: scrape at hours $scrape_hours"
	fi
	time_zone=$(readlink "$destination_root/etc/localtime" 2>/dev/null | sed 's|^.*/zoneinfo/||' || true)
	echo "The schedule runs in this host's local time, which is ${time_zone:-UTC}."
}

print_steps_left_by_hand() {
	cat <<STEPS_LEFT_BY_HAND

This was a first install. These steps are left to do by hand:
  1. Replace $settings_folder_path/www.seek.com.au.json and $settings_folder_path/au.jora.com.json with the real ones:
     doas install -o root -g $service_account -m 0640 <file> $settings_folder_path/
  2. Fill in $settings_folder_path/runtask.json with doas vi, using the password set on runtask@skillsurvey.com.
  3. Run this script again with --start-schedule --scrape-hours HOURS, such as --scrape-hours 7,17.
STEPS_LEFT_BY_HAND
}

refuse_unless_run_as_root
parse_arguments "$@"
refuse_unless_bundle_was_built_for_this_host
refuse_unless_bundle_files_match_their_checksums
refuse_unless_existing_account_is_as_expected "$service_account" "$service_login_class" "$service_home_path" /sbin/nologin
refuse_unless_x_libraries_are_present

first_install=no
[ -f "$install_record" ] || first_install=yes

create_login_class_account_and_folders
install_chromium_if_missing
install_programs_and_settings
record_the_installed_bundle
start_schedule_if_asked
[ "$first_install" = no ] || print_steps_left_by_hand
report_change_count
