import unittest
from cgroup import Cgroup


class CgroupTest(unittest.TestCase):
    def test_empty_cgroups(self):
        def create():
            pass

        cg = Cgroup([])
        cg.destroy()

    def test_single_cgroup_no_key(self):
        def create(cgroup_types, location):
            self.fail("Did not expect any cgroups to be created")

        def destroy(cgroup_types, location):
            self.fail("Did not expect any cgroups to be destroyed")

        def set_parameter(cgroup_type, location, parameter, value):
            self.fail("Did not expect any parameters to be set")

        cg = Cgroup(["/A"], create, destroy, set_parameter)

    def test_single_cgroup_no_type(self):
        def create(cgroup_types, location):
            self.fail("Did not expect any cgroups to be created")

        def destroy(cgroup_types, location):
            self.fail("Did not expect any cgroups to be destroyed")

        def set_parameter(cgroup_type, location, parameter, value):
            self.fail("Did not expect any parameters to be set")

        cg = Cgroup(["/A/bar=1"], create, destroy, set_parameter)

    def test_single_cgroup_no_value(self):
        def create(cgroup_types, location):
            self.fail("Did not expect any cgroups to be created")

        def destroy(cgroup_types, location):
            self.fail("Did not expect any cgroups to be destroyed")

        def set_parameter(cgroup_type, location, parameter, value):
            self.fail("Did not expect any parameters to be set")

        cg = Cgroup(["/A/bar.foo"], create, destroy, set_parameter)
        cg = Cgroup(["/A/foo.bar="], create, destroy, set_parameter)

    def test_single_cgroup(self):
        def create(cgroup_types, location):
            if not hasattr(create, "called"):
                create.called = 0
            create.called += 1

            self.assertEqual(cgroup_types, ["bar"])
            self.assertEqual(location, "/A")

            return True

        def destroy(cgroup_types, location):
            if not hasattr(destroy, "called"):
                destroy.called = 0
            destroy.called += 1

            self.assertEqual(cgroup_types, ["bar"])
            self.assertEqual(location, "/A")

            return True

        def set_parameter(cgroup_type, location, parameter, value):
            if not hasattr(set_parameter, "called"):
                set_parameter.called = 0
            set_parameter.called += 1

            self.assertEqual(cgroup_type, "bar")
            self.assertEqual(location, "/A")
            self.assertEqual(parameter, "bar.foo")
            self.assertEqual(value, "1")

            return True

        cg = Cgroup(["/A/bar.foo=1"], create, destroy, set_parameter)
        cg.destroy()

        # Ensure at most one cgroup is created and destroyed and that at most one parameter is set.
        self.assertEqual(create.called, 1)
        self.assertEqual(set_parameter.called, 1)
        self.assertEqual(destroy.called, 1)

    def test_nested_cgroup(self):
        def create(cgroup_types, location):
            if not hasattr(create, "called"):
                create.called = 0
            create.called += 1

            if create.called == 1:
                # First invocation must be the parent.
                self.assertEqual(cgroup_types, ["bar"])
                self.assertEqual(location, "/A")
            elif create.called == 2:
                # Second invocation must be the child.
                self.assertEqual(cgroup_types, ["bar"])
                self.assertEqual(location, "/A/B")

            return True

        def destroy(cgroup_types, location):
            if not hasattr(destroy, "called"):
                destroy.called = 0
            destroy.called += 1

            if destroy.called == 1:
                # First invocation must be the child.
                self.assertEqual(cgroup_types, ["bar"])
                self.assertEqual(location, "/A/B")
            elif destroy.called == 2:
                # Second invocation must be the parent.
                self.assertEqual(cgroup_types, ["bar"])
                self.assertEqual(location, "/A")

            return True

        def set_parameter(cgroup_type, location, parameter, value):
            if not hasattr(set_parameter, "called"):
                set_parameter.called = 0
            set_parameter.called += 1

            self.assertEqual(cgroup_type, "bar")
            self.assertEqual(location, "/A/B")
            self.assertEqual(parameter, "bar.foo")
            self.assertEqual(value, "1")

            return True

        cg = Cgroup(["/A/B/bar.foo=1"], create, destroy, set_parameter)
        cg.destroy()

        self.assertEqual(create.called, 2)
        self.assertEqual(set_parameter.called, 1)
        self.assertEqual(destroy.called, 2)

    def test_nested_cgroup_with_parent_setting(self):
        def create(cgroup_types, location):
            if not hasattr(create, "called"):
                create.called = 0
            create.called += 1

            if create.called == 1:
                # First invocation must be the parent.
                self.assertEqual(cgroup_types, ["bar"])
                self.assertEqual(location, "/A")
            elif create.called == 2:
                # Second invocation must be the child.
                self.assertEqual(cgroup_types, ["bar"])
                self.assertEqual(location, "/A/B")

            return True

        def destroy(cgroup_types, location):
            if not hasattr(destroy, "called"):
                destroy.called = 0
            destroy.called += 1

            if destroy.called == 1:
                # First invocation must be the child.
                self.assertEqual(cgroup_types, ["bar"])
                self.assertEqual(location, "/A/B")
            elif destroy.called == 2:
                # Second invocation must be the parent.
                self.assertEqual(cgroup_types, ["bar"])
                self.assertEqual(location, "/A")

            return True

        def set_parameter(cgroup_type, location, parameter, value):
            if not hasattr(set_parameter, "called"):
                set_parameter.called = 0
            set_parameter.called += 1

            if set_parameter.called == 1:
                # Parent parameter must be set first
                self.assertEqual(cgroup_type, "bar")
                self.assertEqual(location, "/A")
                self.assertEqual(parameter, "bar.baz")
                self.assertEqual(value, "2")
            elif set_parameter.called == 2:
                self.assertEqual(cgroup_type, "bar")
                self.assertEqual(location, "/A/B")
                self.assertEqual(parameter, "bar.foo")
                self.assertEqual(value, "1")

            return True

        cg = Cgroup([
            "/A/bar.baz=2",
            "/A/B/bar.foo=1"
        ], create, destroy, set_parameter)
        cg.destroy()

        self.assertEqual(create.called, 2)
        self.assertEqual(set_parameter.called, 2)
        self.assertEqual(destroy.called, 2)

    def test_parameter_order(self):
        def create(cgroup_types, location):
            if not hasattr(create, "called"):
                create.called = 0
            create.called += 1

            return True

        def destroy(cgroup_types, location):
            if not hasattr(destroy, "called"):
                destroy.called = 0
            destroy.called += 1

            return True

        def set_parameter(cgroup_type, location, parameter, value):
            if not hasattr(set_parameter, "called"):
                set_parameter.called = 0
            set_parameter.called += 1

            if set_parameter.called == 1:
                # Parent parameter must be set first
                self.assertEqual(cgroup_type, "bar")
                self.assertEqual(location, "/A")
                self.assertEqual(parameter, "bar.qux")
                self.assertEqual(value, "3")
            elif set_parameter.called == 2:
                self.assertEqual(cgroup_type, "bar")
                self.assertEqual(location, "/A")
                self.assertEqual(parameter, "bar.baz")
                self.assertEqual(value, "2")

            return True

        cg = Cgroup([
            "/A/bar.qux=3",
            "/A/bar.baz=2"
        ], create, destroy, set_parameter)
        cg.destroy()

        self.assertEqual(create.called, 1)
        self.assertEqual(set_parameter.called, 2)
        self.assertEqual(destroy.called, 1)
