import glog as log
import subprocess
import sys


class Cgroup:
    def __init__(self, desired_states):
        self.desired_states = desired_states

        self.cgroup_types = {}
        self.hierachies = []

        hierarchy_tree = {}

        for desired_state in desired_states:
            components = desired_state.split("=")
            state_value = components[1]

            if len(components) != 2:
                continue

            hierarchy_components = components[0].split("/")

            # Remove leading component if empty (will happen if path is prefixed with '/')
            if hierarchy_components[0] == "":
                hierarchy = hierarchy_components[1:len(hierarchy_components) - 1]
            else:
                hierarchy = hierarchy_components[0:len(hierarchy_components) - 1]

            cgroup_type = hierarchy_components[len(hierarchy_components) - 1]

            cgroup_type_components = cgroup_type.split('.')
            self.cgroup_types[cgroup_type_components[0]] = cgroup_type_components[0]

            # Build hierarchy tree
            root = hierarchy_tree
            for path_component in hierarchy:
                if path_component not in root:
                    root[path_component] = {}

                root = root[path_component]

                # If at leaf, add cgroup setting.
                if path_component is hierarchy[-1]:
                    root[cgroup_type] = state_value

        # Create cgroups
        # Depth first pre-order traversal of tree
        def dfs(root, location):
            # Create cgroup
            # NOTE: Skip first (empty) root
            if location is not '':
                log.info("Creating cgroup " + ",".join(self.cgroup_types) + ':' + location)
                if subprocess.call(["cgcreate", "-g", ",".join(self.cgroup_types) + ':' + location]) is not 0:
                    log.fatal("Could not create cgroup")
                    sys.exit(1)

            for key, element in root.iteritems():
                if '.' in key:
                    parameter_component = key.split('.')
                    parameter_type = parameter_component[0]
            
                    print(["sh", "-c", "echo '%s' > /sys/fs/cgroup/%s%s/%s" % (element, parameter_type, location, key)])

                    if subprocess.call(["sh", "-c", "echo '%s' > /sys/fs/cgroup/%s%s/%s" % (element, parameter_type, location, key)]) is not 0:
                        log.fatal("Could not set configuration: %s = %s" % (key, element))
                        sys.exit(1)
                else:
                    dfs(element, "/".join([location, key]))

        dfs(hierarchy_tree, '')
        self.hierarchy_tree = hierarchy_tree

    def execute(self, location, command):
        return " ".join(["cgexec", "-g", ",".join(self.cgroup_types) + ":" + location, str(command)])

    def destroy(self):
        # Depth first post-order traversal of tree
        def dfs(root, location):
            for key, element in root.iteritems():
                if '.' not in key:
                    dfs(element, "/".join([location, key]))

            if location is not '':
                if subprocess.call(['cgdelete', ",".join(self.cgroup_types) + ":" + location]) is not 0:
                    log.fatal('could not delete cgroup: ' + location)
                    sys.exit(1)

        dfs(self.hierarchy_tree, '')
