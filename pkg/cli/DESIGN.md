I see value in separation of ENVIRONMENT and PARAMETERS.
When I create an Experiment I want to determine that my Memcached runs on 4 threads.
That's my experiment property.

But there are some variables that are above those properties.
Values like Cassandra location depends on Enviroment. I cannot foresee their location
 while writing experiment. But I want my experiment binary to obtain those values
 from respective ENVIROMENT.

I want my exp binary to work both on my DEVELOPER, STAGING and PROD enviroments.
I can have my environments prepared by Ansible scripts,
that can be checked-in to our sno-ops repository. This way our ENVIROMENTS are the same.

For me, there are some crucial aspects of software development:

- Keep code entanglement as low as possible
- Have one single way of defining something
- Keep responsibilities encapsulated in their respectable domains


In my code example we have following properties:
- CLI does not know anything about Experiment and Launcher
- Launcher dos not know anything about CLI
- Every Launcher can tell what he requires from environment (responsibility encapsulation)
- Experiment is a proxy that passes Launchers env reqs to CLI (law of demeter)
- Only Launcher can configure it's own environment configuration - Experiment should not do that, because it's responsibilites (experiment cannot know where Cassandra will be Launched in env) (responsibility segregation)
- There is only one single way to set environment configuration in launchers (simple code, less bugs)

IMO opinion, an approach where Experiment defines what Launcher needs is wrong. It look fragile because it's easy to overlook some variables. It can also cause regression problems when Launcher would define NEW env requirements (we would have to change all the experiments to accommodate this).

Also, I see problem when Experiment even touches Launchers env configuration. It cannot know it beforehand, so why he should do it? During our discussions, Bartek proposed similar thing which is IMO very good idea.


FAQ:
-

1. Why we are even discussing such small thing?

I don't consider it small. Entangling CLI with Experiment and Launcher brings much overhead to whole code. I consider this a core problem.

2. How is this different from other propositions?

In this PR I have created a CLI that is powerful and simple.
With IOC by EnvVariables and simple os.setenv() in CLI we can do anything.
Helper method gives great readme for experiment user and it is very close to entity that utilizes variables (Launchers defines env reqs AND helper).
