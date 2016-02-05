import glog as log
import subprocess
import sys
import time
import collections


class Cgroup:
    """
    Cgroup hierarchy control class: creates and configures cgroups hierarchies based on a set of 'desired states'.
    See __init__() documentation for format of desired states.
    """

    def __init__(self, desired_states, create_func=None, destroy_func=None, set_func=None):
        """
        :param desired_states: List of desired states in form of <path>=<value>.
                               For example, ["/A/B/cpu.shares=1", "/A/C/mem.limit_in_bytes=512"]
                               The above will create:
                               1) The 'A' root of the two children with 'cpu' and 'mem' cgroups
                               2) 'B' child cgroup with 'cpu' and 'mem' and set cpu shares to 1
                               3) 'C' child cgroup with 'cpu' and 'mem' and set memory limit to 512 bytes

                               The settings will be applied in order, so settings in the parents can be set before the
                               children's settings are applied. This is necessary for setting 'cpuset's.
        :param create_func:    Function to use for creating cgroups. For testing purposes only.
        :param destroy_func:   Function to use for destroy cgroups. For testing purposes only.
        :param set_func:       Function to use for setting cgroup parameters. For testing purposes only.
        """
        self.desired_states = desired_states
        self.cgroup_types = collections.OrderedDict()
        self.hierarchies = []

        # For testing purposes, creating, destroying and changing cgroups settings have been pulled into functions
        # which can be provided by the caller.
        def create(cgroup_types, location):
            """
            :param cgroup_types: List of types. For example ['cpu', 'mem']
            :param location: Path of cgroup (relative from '/sys/fs/cgroup/<type>').
                   For example, '/A/B' for '/sys/fs/cgroup/cpu/A/B'.
            :return: True if operation succeeded.
            """
            log.info("creating cgroup " + ",".join(cgroup_types) + ":" + location)
            if subprocess.call(["cgcreate", "-g", ",".join(self.cgroup_types) + ":" + location]) is not 0:
                return False
            return True

        def destroy(cgroup_types, location):
            """
            :param cgroup_types: List of types. For example ['cpu', 'mem']
            :param location:     Path of cgroup (relative from '/sys/fs/cgroup/<type>').
                                 For example, '/A/B' for '/sys/fs/cgroup/cpu/A/B'.
            :return:             True if operation succeeded.
            """
            if subprocess.call(["cgdelete", ",".join(cgroup_types) + ":" + location]) is not 0:
                return False
            return True

        def set_parameter(cgroup_type, location, parameter, value):
            """
            :param cgroup_type:  Cgroup type. For example 'cpu' or 'mem'
            :param location:     Path of cgroup (relative from '/sys/fs/cgroup/<type>').
                                 For example, '/A/B' for '/sys/fs/cgroup/cpu/A/B'.
            :param parameter:    Parameter to set. For example 'cpu.shares' for '/A/B/cpu.shares=1'
            :param value:        Value to set. For example '1' for '/A/B/cpu.shares=1'
            :return:             True if operation succeeded.
            """
            if subprocess.call(["sh", "-c", "echo '%s' > /sys/fs/cgroup/%s%s/%s" % (value, cgroup_type, location, parameter)]) is not 0:
                return False
            return True

        if create_func is None:
            create_func = create

        if destroy_func is None:
            destroy_func = destroy

        if set_func is None:
            set_func = set_parameter

        self.create_func = create_func
        self.destroy_func = destroy_func
        self.set_func = set_func

        hierarchy_tree = collections.OrderedDict()

        for desired_state in desired_states:
            components = desired_state.split("=")

            if len(components) != 2:
                log.warning("unknown format: '%s'. skipping!" % desired_state)
                continue

            state_value = components[1]

            if state_value == "":
                log.warning("no value set in: '%s'. skipping!" % desired_state)
                continue

            hierarchy_components = components[0].split("/")

            # Remove leading component if empty (will happen if path is prefixed with '/')
            if hierarchy_components[0] == "":
                hierarchy = hierarchy_components[1:len(hierarchy_components) - 1]
            else:
                hierarchy = hierarchy_components[0:len(hierarchy_components) - 1]

            cgroup_type = hierarchy_components[len(hierarchy_components) - 1]

            cgroup_type_components = cgroup_type.split(".")

            if len(cgroup_type_components) != 2:
                log.warning("Could not determine cgroup type from parameter name '%s'. skipping!")
                continue

            self.cgroup_types[cgroup_type_components[0]] = cgroup_type_components[0]

            # Build hierarchy tree
            root = hierarchy_tree
            for path_component in hierarchy:
                if path_component not in root:
                    root[path_component] = collections.OrderedDict()

                root = root[path_component]

                # If at leaf, add cgroup setting.
                if path_component is hierarchy[-1]:
                    root[cgroup_type] = state_value

        # Create cgroups
        # Depth first pre-order traversal of tree
        def dfs(root, location):
            # Create cgroup
            # NOTE: Skip first (empty) root
            if location is not "":
                log.info("creating cgroup " + ",".join(self.cgroup_types) + ":" + location)
                if not self.create_func(",".join(self.cgroup_types), location):
                    log.fatal("could not create cgroup: %s" % location)
                    sys.exit(1)

            # NOTE: We have to apply parent settings before child setting in cgroups like "cpuset".
            for key, element in root.iteritems():
                if "." in key:
                    parameter_component = key.split(".")
                    parameter_type = parameter_component[0]

                    log.info("setting %s=%s" % (location + "/" + key, element))

                    if not self.set_func(parameter_type, location, key, element):
                        log.fatal("Could not set configuration: %s = %s" % (key, element))
                        sys.exit(1)

            for key, element in root.iteritems():
                if "." not in key:
                    dfs(element, "/".join([location, key]))

        dfs(hierarchy_tree, "")
        self.hierarchy_tree = hierarchy_tree

    def execute(self, location, command):
        return " ".join(["cgexec", "-g", ",".join(self.cgroup_types) + ":" + location, str(command)])

    def destroy(self):
        # Depth first post-order traversal of tree
        def dfs(root, location):
            for key, element in root.iteritems():
                if "." not in key:
                    dfs(element, "/".join([location, key]))

            if location is not "":
                log.info("deleting cgroup " + ",".join(self.cgroup_types) + ":" + location)

                if not self.destroy_func(",".join(self.cgroup_types), location):
                    log.fatal("could not delete cgroup: " + location)
                    sys.exit(1)

        dfs(self.hierarchy_tree, "")

        # HACK: Make sure cgroups deletion is done before recreating exlusive cpusets (which will otherwise fail).
        # TODO(nnielsen): Add more deterministic way of telling that the operation is done.
        time.sleep(0.5)
