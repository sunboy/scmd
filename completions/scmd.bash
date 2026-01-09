# bash completion for scmd                                 -*- shell-script -*-

__scmd_debug()
{
    if [[ -n ${BASH_COMP_DEBUG_FILE:-} ]]; then
        echo "$*" >> "${BASH_COMP_DEBUG_FILE}"
    fi
}

# Homebrew on Macs have version 1.3 of bash-completion which doesn't include
# _init_completion. This is a very minimal version of that function.
__scmd_init_completion()
{
    COMPREPLY=()
    _get_comp_words_by_ref "$@" cur prev words cword
}

__scmd_index_of_word()
{
    local w word=$1
    shift
    index=0
    for w in "$@"; do
        [[ $w = "$word" ]] && return
        index=$((index+1))
    done
    index=-1
}

__scmd_contains_word()
{
    local w word=$1; shift
    for w in "$@"; do
        [[ $w = "$word" ]] && return
    done
    return 1
}

__scmd_handle_go_custom_completion()
{
    __scmd_debug "${FUNCNAME[0]}: cur is ${cur}, words[*] is ${words[*]}, #words[@] is ${#words[@]}"

    local shellCompDirectiveError=1
    local shellCompDirectiveNoSpace=2
    local shellCompDirectiveNoFileComp=4
    local shellCompDirectiveFilterFileExt=8
    local shellCompDirectiveFilterDirs=16

    local out requestComp lastParam lastChar comp directive args

    # Prepare the command to request completions for the program.
    # Calling ${words[0]} instead of directly scmd allows handling aliases
    args=("${words[@]:1}")
    # Disable ActiveHelp which is not supported for bash completion v1
    requestComp="SCMD_ACTIVE_HELP=0 ${words[0]} __completeNoDesc ${args[*]}"

    lastParam=${words[$((${#words[@]}-1))]}
    lastChar=${lastParam:$((${#lastParam}-1)):1}
    __scmd_debug "${FUNCNAME[0]}: lastParam ${lastParam}, lastChar ${lastChar}"

    if [ -z "${cur}" ] && [ "${lastChar}" != "=" ]; then
        # If the last parameter is complete (there is a space following it)
        # We add an extra empty parameter so we can indicate this to the go method.
        __scmd_debug "${FUNCNAME[0]}: Adding extra empty parameter"
        requestComp="${requestComp} \"\""
    fi

    __scmd_debug "${FUNCNAME[0]}: calling ${requestComp}"
    # Use eval to handle any environment variables and such
    out=$(eval "${requestComp}" 2>/dev/null)

    # Extract the directive integer at the very end of the output following a colon (:)
    directive=${out##*:}
    # Remove the directive
    out=${out%:*}
    if [ "${directive}" = "${out}" ]; then
        # There is not directive specified
        directive=0
    fi
    __scmd_debug "${FUNCNAME[0]}: the completion directive is: ${directive}"
    __scmd_debug "${FUNCNAME[0]}: the completions are: ${out}"

    if [ $((directive & shellCompDirectiveError)) -ne 0 ]; then
        # Error code.  No completion.
        __scmd_debug "${FUNCNAME[0]}: received error from custom completion go code"
        return
    else
        if [ $((directive & shellCompDirectiveNoSpace)) -ne 0 ]; then
            if [[ $(type -t compopt) = "builtin" ]]; then
                __scmd_debug "${FUNCNAME[0]}: activating no space"
                compopt -o nospace
            fi
        fi
        if [ $((directive & shellCompDirectiveNoFileComp)) -ne 0 ]; then
            if [[ $(type -t compopt) = "builtin" ]]; then
                __scmd_debug "${FUNCNAME[0]}: activating no file completion"
                compopt +o default
            fi
        fi
    fi

    if [ $((directive & shellCompDirectiveFilterFileExt)) -ne 0 ]; then
        # File extension filtering
        local fullFilter filter filteringCmd
        # Do not use quotes around the $out variable or else newline
        # characters will be kept.
        for filter in ${out}; do
            fullFilter+="$filter|"
        done

        filteringCmd="_filedir $fullFilter"
        __scmd_debug "File filtering command: $filteringCmd"
        $filteringCmd
    elif [ $((directive & shellCompDirectiveFilterDirs)) -ne 0 ]; then
        # File completion for directories only
        local subdir
        # Use printf to strip any trailing newline
        subdir=$(printf "%s" "${out}")
        if [ -n "$subdir" ]; then
            __scmd_debug "Listing directories in $subdir"
            __scmd_handle_subdirs_in_dir_flag "$subdir"
        else
            __scmd_debug "Listing directories in ."
            _filedir -d
        fi
    else
        while IFS='' read -r comp; do
            COMPREPLY+=("$comp")
        done < <(compgen -W "${out}" -- "$cur")
    fi
}

__scmd_handle_reply()
{
    __scmd_debug "${FUNCNAME[0]}"
    local comp
    case $cur in
        -*)
            if [[ $(type -t compopt) = "builtin" ]]; then
                compopt -o nospace
            fi
            local allflags
            if [ ${#must_have_one_flag[@]} -ne 0 ]; then
                allflags=("${must_have_one_flag[@]}")
            else
                allflags=("${flags[*]} ${two_word_flags[*]}")
            fi
            while IFS='' read -r comp; do
                COMPREPLY+=("$comp")
            done < <(compgen -W "${allflags[*]}" -- "$cur")
            if [[ $(type -t compopt) = "builtin" ]]; then
                [[ "${COMPREPLY[0]}" == *= ]] || compopt +o nospace
            fi

            # complete after --flag=abc
            if [[ $cur == *=* ]]; then
                if [[ $(type -t compopt) = "builtin" ]]; then
                    compopt +o nospace
                fi

                local index flag
                flag="${cur%=*}"
                __scmd_index_of_word "${flag}" "${flags_with_completion[@]}"
                COMPREPLY=()
                if [[ ${index} -ge 0 ]]; then
                    PREFIX=""
                    cur="${cur#*=}"
                    ${flags_completion[${index}]}
                    if [ -n "${ZSH_VERSION:-}" ]; then
                        # zsh completion needs --flag= prefix
                        eval "COMPREPLY=( \"\${COMPREPLY[@]/#/${flag}=}\" )"
                    fi
                fi
            fi

            if [[ -z "${flag_parsing_disabled}" ]]; then
                # If flag parsing is enabled, we have completed the flags and can return.
                # If flag parsing is disabled, we may not know all (or any) of the flags, so we fallthrough
                # to possibly call handle_go_custom_completion.
                return 0;
            fi
            ;;
    esac

    # check if we are handling a flag with special work handling
    local index
    __scmd_index_of_word "${prev}" "${flags_with_completion[@]}"
    if [[ ${index} -ge 0 ]]; then
        ${flags_completion[${index}]}
        return
    fi

    # we are parsing a flag and don't have a special handler, no completion
    if [[ ${cur} != "${words[cword]}" ]]; then
        return
    fi

    local completions
    completions=("${commands[@]}")
    if [[ ${#must_have_one_noun[@]} -ne 0 ]]; then
        completions+=("${must_have_one_noun[@]}")
    elif [[ -n "${has_completion_function}" ]]; then
        # if a go completion function is provided, defer to that function
        __scmd_handle_go_custom_completion
    fi
    if [[ ${#must_have_one_flag[@]} -ne 0 ]]; then
        completions+=("${must_have_one_flag[@]}")
    fi
    while IFS='' read -r comp; do
        COMPREPLY+=("$comp")
    done < <(compgen -W "${completions[*]}" -- "$cur")

    if [[ ${#COMPREPLY[@]} -eq 0 && ${#noun_aliases[@]} -gt 0 && ${#must_have_one_noun[@]} -ne 0 ]]; then
        while IFS='' read -r comp; do
            COMPREPLY+=("$comp")
        done < <(compgen -W "${noun_aliases[*]}" -- "$cur")
    fi

    if [[ ${#COMPREPLY[@]} -eq 0 ]]; then
        if declare -F __scmd_custom_func >/dev/null; then
            # try command name qualified custom func
            __scmd_custom_func
        else
            # otherwise fall back to unqualified for compatibility
            declare -F __custom_func >/dev/null && __custom_func
        fi
    fi

    # available in bash-completion >= 2, not always present on macOS
    if declare -F __ltrim_colon_completions >/dev/null; then
        __ltrim_colon_completions "$cur"
    fi

    # If there is only 1 completion and it is a flag with an = it will be completed
    # but we don't want a space after the =
    if [[ "${#COMPREPLY[@]}" -eq "1" ]] && [[ $(type -t compopt) = "builtin" ]] && [[ "${COMPREPLY[0]}" == --*= ]]; then
       compopt -o nospace
    fi
}

# The arguments should be in the form "ext1|ext2|extn"
__scmd_handle_filename_extension_flag()
{
    local ext="$1"
    _filedir "@(${ext})"
}

__scmd_handle_subdirs_in_dir_flag()
{
    local dir="$1"
    pushd "${dir}" >/dev/null 2>&1 && _filedir -d && popd >/dev/null 2>&1 || return
}

__scmd_handle_flag()
{
    __scmd_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    # if a command required a flag, and we found it, unset must_have_one_flag()
    local flagname=${words[c]}
    local flagvalue=""
    # if the word contained an =
    if [[ ${words[c]} == *"="* ]]; then
        flagvalue=${flagname#*=} # take in as flagvalue after the =
        flagname=${flagname%=*} # strip everything after the =
        flagname="${flagname}=" # but put the = back
    fi
    __scmd_debug "${FUNCNAME[0]}: looking for ${flagname}"
    if __scmd_contains_word "${flagname}" "${must_have_one_flag[@]}"; then
        must_have_one_flag=()
    fi

    # if you set a flag which only applies to this command, don't show subcommands
    if __scmd_contains_word "${flagname}" "${local_nonpersistent_flags[@]}"; then
      commands=()
    fi

    # keep flag value with flagname as flaghash
    # flaghash variable is an associative array which is only supported in bash > 3.
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        if [ -n "${flagvalue}" ] ; then
            flaghash[${flagname}]=${flagvalue}
        elif [ -n "${words[ $((c+1)) ]}" ] ; then
            flaghash[${flagname}]=${words[ $((c+1)) ]}
        else
            flaghash[${flagname}]="true" # pad "true" for bool flag
        fi
    fi

    # skip the argument to a two word flag
    if [[ ${words[c]} != *"="* ]] && __scmd_contains_word "${words[c]}" "${two_word_flags[@]}"; then
        __scmd_debug "${FUNCNAME[0]}: found a flag ${words[c]}, skip the next argument"
        c=$((c+1))
        # if we are looking for a flags value, don't show commands
        if [[ $c -eq $cword ]]; then
            commands=()
        fi
    fi

    c=$((c+1))

}

__scmd_handle_noun()
{
    __scmd_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    if __scmd_contains_word "${words[c]}" "${must_have_one_noun[@]}"; then
        must_have_one_noun=()
    elif __scmd_contains_word "${words[c]}" "${noun_aliases[@]}"; then
        must_have_one_noun=()
    fi

    nouns+=("${words[c]}")
    c=$((c+1))
}

__scmd_handle_command()
{
    __scmd_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    local next_command
    if [[ -n ${last_command} ]]; then
        next_command="_${last_command}_${words[c]//:/__}"
    else
        if [[ $c -eq 0 ]]; then
            next_command="_scmd_root_command"
        else
            next_command="_${words[c]//:/__}"
        fi
    fi
    c=$((c+1))
    __scmd_debug "${FUNCNAME[0]}: looking for ${next_command}"
    declare -F "$next_command" >/dev/null && $next_command
}

__scmd_handle_word()
{
    if [[ $c -ge $cword ]]; then
        __scmd_handle_reply
        return
    fi
    __scmd_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"
    if [[ "${words[c]}" == -* ]]; then
        __scmd_handle_flag
    elif __scmd_contains_word "${words[c]}" "${commands[@]}"; then
        __scmd_handle_command
    elif [[ $c -eq 0 ]]; then
        __scmd_handle_command
    elif __scmd_contains_word "${words[c]}" "${command_aliases[@]}"; then
        # aliashash variable is an associative array which is only supported in bash > 3.
        if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
            words[c]=${aliashash[${words[c]}]}
            __scmd_handle_command
        else
            __scmd_handle_noun
        fi
    else
        __scmd_handle_noun
    fi
    __scmd_handle_word
}

_scmd_backends()
{
    last_command="scmd_backends"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_cache_clear()
{
    last_command="scmd_cache_clear"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_cache_stats()
{
    last_command="scmd_cache_stats"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_cache()
{
    last_command="scmd_cache"

    command_aliases=()

    commands=()
    commands+=("clear")
    commands+=("stats")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_completion()
{
    last_command="scmd_completion"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--help")
    flags+=("-h")
    local_nonpersistent_flags+=("--help")
    local_nonpersistent_flags+=("-h")
    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    must_have_one_noun+=("bash")
    must_have_one_noun+=("fish")
    must_have_one_noun+=("powershell")
    must_have_one_noun+=("zsh")
    noun_aliases=()
}

_scmd_config()
{
    last_command="scmd_config"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_doctor()
{
    last_command="scmd_doctor"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_explain()
{
    last_command="scmd_explain"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_help()
{
    last_command="scmd_help"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    has_completion_function=1
    noun_aliases=()
}

_scmd_kill-process()
{
    last_command="scmd_kill-process"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_lock_generate()
{
    last_command="scmd_lock_generate"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--output")
    local_nonpersistent_flags+=("--output=")
    local_nonpersistent_flags+=("-o")
    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_lock_install()
{
    last_command="scmd_lock_install"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_lock()
{
    last_command="scmd_lock"

    command_aliases=()

    commands=()
    commands+=("generate")
    commands+=("install")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_models_default()
{
    last_command="scmd_models_default"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_models_info()
{
    last_command="scmd_models_info"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_models_list()
{
    last_command="scmd_models_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_models_pull()
{
    last_command="scmd_models_pull"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_models_remove()
{
    last_command="scmd_models_remove"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_models()
{
    last_command="scmd_models"

    command_aliases=()

    commands=()
    commands+=("default")
    commands+=("info")
    commands+=("list")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("ls")
        aliashash["ls"]="list"
    fi
    commands+=("pull")
    commands+=("remove")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("delete")
        aliashash["delete"]="remove"
        command_aliases+=("rm")
        aliashash["rm"]="remove"
    fi

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_registry_categories()
{
    last_command="scmd_registry_categories"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_registry_featured()
{
    last_command="scmd_registry_featured"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_registry_search()
{
    last_command="scmd_registry_search"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--category=")
    two_word_flags+=("--category")
    local_nonpersistent_flags+=("--category")
    local_nonpersistent_flags+=("--category=")
    flags+=("--featured")
    local_nonpersistent_flags+=("--featured")
    flags+=("--limit=")
    two_word_flags+=("--limit")
    local_nonpersistent_flags+=("--limit")
    local_nonpersistent_flags+=("--limit=")
    flags+=("--sort=")
    two_word_flags+=("--sort")
    local_nonpersistent_flags+=("--sort")
    local_nonpersistent_flags+=("--sort=")
    flags+=("--verified")
    local_nonpersistent_flags+=("--verified")
    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_registry()
{
    last_command="scmd_registry"

    command_aliases=()

    commands=()
    commands+=("categories")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("cats")
        aliashash["cats"]="categories"
    fi
    commands+=("featured")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("popular")
        aliashash["popular"]="featured"
        command_aliases+=("trending")
        aliashash["trending"]="featured"
    fi
    commands+=("search")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_repo_add()
{
    last_command="scmd_repo_add"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_repo_install()
{
    last_command="scmd_repo_install"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_repo_list()
{
    last_command="scmd_repo_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_repo_remove()
{
    last_command="scmd_repo_remove"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_repo_search()
{
    last_command="scmd_repo_search"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_repo_show()
{
    last_command="scmd_repo_show"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_repo_update()
{
    last_command="scmd_repo_update"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_repo()
{
    last_command="scmd_repo"

    command_aliases=()

    commands=()
    commands+=("add")
    commands+=("install")
    commands+=("list")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("ls")
        aliashash["ls"]="list"
    fi
    commands+=("remove")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("delete")
        aliashash["delete"]="remove"
        command_aliases+=("rm")
        aliashash["rm"]="remove"
    fi
    commands+=("search")
    commands+=("show")
    commands+=("update")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_review()
{
    last_command="scmd_review"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_server_logs()
{
    last_command="scmd_server_logs"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--tail=")
    two_word_flags+=("--tail")
    local_nonpersistent_flags+=("--tail")
    local_nonpersistent_flags+=("--tail=")
    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_server_restart()
{
    last_command="scmd_server_restart"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_server_start()
{
    last_command="scmd_server_start"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    local_nonpersistent_flags+=("--context")
    local_nonpersistent_flags+=("--context=")
    local_nonpersistent_flags+=("-c")
    flags+=("--cpu")
    local_nonpersistent_flags+=("--cpu")
    flags+=("--gpu")
    local_nonpersistent_flags+=("--gpu")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    local_nonpersistent_flags+=("--model")
    local_nonpersistent_flags+=("--model=")
    local_nonpersistent_flags+=("-m")
    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_server_status()
{
    last_command="scmd_server_status"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_server_stop()
{
    last_command="scmd_server_stop"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_server()
{
    last_command="scmd_server"

    command_aliases=()

    commands=()
    commands+=("logs")
    commands+=("restart")
    commands+=("start")
    commands+=("status")
    commands+=("stop")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_setup()
{
    last_command="scmd_setup"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--force")
    local_nonpersistent_flags+=("--force")
    flags+=("--quiet")
    local_nonpersistent_flags+=("--quiet")
    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_slash_add()
{
    last_command="scmd_slash_add"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--alias=")
    two_word_flags+=("--alias")
    local_nonpersistent_flags+=("--alias")
    local_nonpersistent_flags+=("--alias=")
    flags+=("--description=")
    two_word_flags+=("--description")
    local_nonpersistent_flags+=("--description")
    local_nonpersistent_flags+=("--description=")
    flags+=("--stdin")
    local_nonpersistent_flags+=("--stdin")
    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_slash_alias()
{
    last_command="scmd_slash_alias"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_slash_init()
{
    last_command="scmd_slash_init"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_slash_interactive()
{
    last_command="scmd_slash_interactive"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_slash_list()
{
    last_command="scmd_slash_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_slash_remove()
{
    last_command="scmd_slash_remove"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_slash_run()
{
    last_command="scmd_slash_run"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_slash()
{
    last_command="scmd_slash"

    command_aliases=()

    commands=()
    commands+=("add")
    commands+=("alias")
    commands+=("init")
    commands+=("interactive")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("i")
        aliashash["i"]="interactive"
        command_aliases+=("repl")
        aliashash["repl"]="interactive"
    fi
    commands+=("list")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("ls")
        aliashash["ls"]="list"
    fi
    commands+=("remove")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("delete")
        aliashash["delete"]="remove"
        command_aliases+=("rm")
        aliashash["rm"]="remove"
    fi
    commands+=("run")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_update()
{
    last_command="scmd_update"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--all")
    local_nonpersistent_flags+=("--all")
    flags+=("--check")
    local_nonpersistent_flags+=("--check")
    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_version()
{
    last_command="scmd_version"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_scmd_root_command()
{
    last_command="scmd"

    command_aliases=()

    commands=()
    commands+=("backends")
    commands+=("cache")
    commands+=("completion")
    commands+=("config")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("cfg")
        aliashash["cfg"]="config"
    fi
    commands+=("doctor")
    commands+=("explain")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("e")
        aliashash["e"]="explain"
        command_aliases+=("what")
        aliashash["what"]="explain"
    fi
    commands+=("help")
    commands+=("kill-process")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("killp")
        aliashash["killp"]="kill-process"
        command_aliases+=("kp")
        aliashash["kp"]="kill-process"
    fi
    commands+=("lock")
    commands+=("models")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("model")
        aliashash["model"]="models"
    fi
    commands+=("registry")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("reg")
        aliashash["reg"]="registry"
    fi
    commands+=("repo")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("repos")
        aliashash["repos"]="repo"
    fi
    commands+=("review")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("r")
        aliashash["r"]="review"
    fi
    commands+=("server")
    commands+=("setup")
    commands+=("slash")
    commands+=("update")
    commands+=("version")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backend=")
    two_word_flags+=("--backend")
    two_word_flags+=("-b")
    flags+=("--context=")
    two_word_flags+=("--context")
    two_word_flags+=("-c")
    flags+=("--context-size=")
    two_word_flags+=("--context-size")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    flags+=("--model=")
    two_word_flags+=("--model")
    two_word_flags+=("-m")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    two_word_flags+=("-p")
    flags+=("--quiet")
    flags+=("-q")
    flags+=("--verbose")
    flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

__start_scmd()
{
    local cur prev words cword split
    declare -A flaghash 2>/dev/null || :
    declare -A aliashash 2>/dev/null || :
    if declare -F _init_completion >/dev/null 2>&1; then
        _init_completion -s || return
    else
        __scmd_init_completion -n "=" || return
    fi

    local c=0
    local flag_parsing_disabled=
    local flags=()
    local two_word_flags=()
    local local_nonpersistent_flags=()
    local flags_with_completion=()
    local flags_completion=()
    local commands=("scmd")
    local command_aliases=()
    local must_have_one_flag=()
    local must_have_one_noun=()
    local has_completion_function=""
    local last_command=""
    local nouns=()
    local noun_aliases=()

    __scmd_handle_word
}

if [[ $(type -t compopt) = "builtin" ]]; then
    complete -o default -F __start_scmd scmd
else
    complete -o default -o nospace -F __start_scmd scmd
fi

# ex: ts=4 sw=4 et filetype=sh
