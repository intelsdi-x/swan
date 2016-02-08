# Tuneable Colocation Scoring

Objective metrics for compute system quality.

## Background

There are orthogonal and contradictory priorities involved in operating
applications in a cluster. Here are some common ones:

1. Maximize cluster utilization.
1. Minimize QoS violations.
1. Maximize batch job throughput.
1. Minimize latency.
1. Minimize variance in a series like tail latency (predictable performance.)
1. Maximize availability (minimize downtime.)

The relative priority of these goals varies widely in different situations.
We propose a composite performance measurement that can be tuned by assigning
weights to each of several objective metrics. The result is called a
_colocation score_.

Performance depends on the specific workload mix as well as the system's
hardware capabilities and enabled isolation strategies. A main idea behind
this experimental framework is to hold the workload mix constant, record
baseline performance, and then vary isolation strategies to measure
how the colocation score is affected.

### Definitions

###### Workload (<code>w<sub>i</sub></code>)
A program or set of programs to run on a given system.

###### Workload stream (<code>&Psi;</code>)
A sequence of workloads
<code>&Psi; = &lang;w<sub>1</sub>, ..., w<sub>n</sub>&rang;</code>.

###### Hardware Resource (`r`)
A hardware compute resource, such as hardware threads, system memory,
CPU sockets, memory bandwidth, network egress bandwidth, or cache.

###### Service level indicator (SLI)
An objective measurement of a workload's performance used in service
declarations, such as a service level agreement. Common measurements include
response latency, downtime, or throughput. SLIs for some batch-type
workloads may be only available as a single sample (_execution time_.)
Each SLI value should be accompanied by a local timestamp in order to
correct for irregular sampling intervals.

###### Normalized service level indicator (nSLI)
SLI sample values that are normalized (contained in the closed interval
<code>[0, 1]</code>) and oriented such that higher values represent
better performance.

###### Service level objective (SLO)
Criteria for acceptable values for one SLI for a workload, expressed as
an interval `[min, max]`. "One-handed" intervals can be expressed as
<code>[min, +&infin;)</code> or <code>(-&infin;, max]</code>.

###### Service level agreement (SLA)
An agreement between service provider and service consumer, describing
service levels and the terms under which service is provided.

###### Quality of service violation (QOS violation)
A contiguous subsequence of SLI samples for which the SLI value falls
outside the tolerance specified by its SLO.

###### Colocation
Workloads <code>w<sub>i</sub></code> and <code>w<sub>j</sub></code> are
_colocated_ if they run simultaneously on the same physical machine.

###### Aggressor
For a given workload <code>w<sub>i</sub></code>, aggressors are colocated
workloads.

### Scores

Below, let <code>&Psi;</code> be a workload stream
<code>&lang;w<sub>1</sub>, ..., w<sub>n</sub>&rang;</code>.

and let `R` be a set of resources
<code>{ r<sub>1</sub>, ..., r<sub>n }</sub></code>.

_Notation:_

- `#S`: the cardinality (size of) the set or sequence `S`
- <code>&Sigma;[S]</code>: the sum of each value in `S`
- <code>s &in; S</code>: `s` is an element of `S`

###### Performance delta score (<code>&Delta;P</code>)
A measure of a the performance reduction between two runs of a workload.
Given two sequences of nSLI samples <code>nSLI<sub>1</sub></code> and
<code>nSLI<sub>2</sub></code> for workload <code>w<sub>i</sub></code>, the
performance degradation is simply the difference of the arithmetic means:
<code>
&Delta;P = (&Sigma;[nSLI<sub>1</sub>] / #nSLI<sub>1</sub>) -
(&Sigma;[nSLI<sub>2</sub> / #nSLI<sub>2</sub>)
</code>

Negative degradation values indicate of course that the workload was
_accelerated_ in run two relative to run one.

###### Violation frequency score (<code>V<sub>&nu;</sub></code>)
The violation frequency is a measurement of how often QoS violations
occurred, as a fraction of total SLI samples. Given the subset of
violating samples from the sequence <code>SLI<sub>i</sub></code>
<code>vs = { s | s &in; SLI<sub>i</sub> &and; &not;contains(SLO, s) }</code>
then the violation frequency
<code>V<sub>&nu;</sub> = #vs / #SLI<sub>i</sub></code>

###### Violation severity score (<code>V<sub>s</sub></code>)
The violation severity score for a series of SLI samples estimates the
cumulative effect of QoS violations on end-user experience. We compute
the area that lies under the normalized SLO line and above the normalized
SLI curve, then normalize again against the total area below the `nSLO`
line. Recall for normalized SLI, large values are preferable.
A series containing zero SLO violations has a severity score of zero,
and a series where all `nSLI` values are on the floor (zero) has the
maximum severity score of one.

For a given series
<code>nSLI = &lang;s<sub>1</sub>, ..., s<sub>n</sub>&rang;</code>:

Violation area
<code>A<sub>V</sub> = &Sigma;[{ max(0, nSLO - s<sub>i</sub>) &times;
s<sub>i+1</sub>.time - s<sub>i</sub>.time | s<sub>i</sub>,
s<sub>i+1</sub> &in; nSLI }]</code>

nSLO area <code>A<sub>nSLO</sub> = nSLO &times;
(s<sub>n</sub>.time - s<sub>1</sub>.time)</code>

Violation severity
<code>V<sub>s</sub> = A<sub>V</sub> / A<sub>nSLO</sub></code>.

###### Pressure score (`P`)
Measures to what extent a workload prevents colocated workloads from
effectively utilizing a given shared resource on a system.
Scores for each resource are normalized based on their quantity or capacity.
For example, an impact score of `0.5` for network egress on a physical
interface means that the workload is consumes half of the available
bandwidth. The impact score for an exlusive resource allocated to a workload
(such as an exclusive cpu set) is defined to be `1.0`.

Let `P` be a function of type <code>P: W &rarr; R &rarr; ℝ</code>.

We estimate the utilization of each resource directly or via a
combination of proxy metrics, depending on monitoring features available
on each compute platform. Initial recommendations follow for
high-priority resource types:

1. **Processor:** _cpu time_
1. **Memory bandwidth:** _membw_
1. **Cache:** _LLC miss_

<!--
TODO(CD): Replace these placeholders with better metrics.
-->

###### Sensitivity score (`S`)
A measurement of how greatly a given workload's performance (SLI) is impacted
due to contention for a given resource. Combining this score with impact
scores forms the basis for interference prediction and avoidance.

**TODO...**

<!--
TODO: Re-read regression algorithm from the multiprocessor interference paper.
TODO: Formulate the algorithm in terms of this doc.
TODO: Write down both the offline and online versions of the algorithm.
TODO: Cite source.
-->

###### Isolation score (`I`)
A measurement of how well a system provides predictable performance for a
given workload. As a first pass, this metric is defined as the
_sample variance_ (average squared deviation) of an `nSLI` series.
For completeness, the formula for sample variance given an `nSLI` series
with arithmetic mean <code>s&#772;</code> is:

<code>I = &Sigma;[(s<sub>i</sub> - s&#772;)<sup>2</sup>] / #nSLI</code>

###### Compute work score (`W`)
A measurement of the useful work performed (for example, &mu;ops commited)
over time by a set of workloads.

###### Colocation score (`C`)
A composite measurement that incorporates several aspects of colocation
quality. As a first pass, we can simply normalize a weighted sum of relevant
subfunctions, given weights
<code>&omega;<sub>V<sub>&nu;</sub></sub></code>,
<code>&omega;<sub>I</sub></code>, and
<code>&omega;<sub>W</sub></code>.

<code>C =
&Sigma;[{ &omega;<sub>V<sub>&nu;</sub></sub>(1 -
V<sub>&nu;<sub>w<sub>i</sub></sub></sub>) +
&omega;<sub>I</sub>I<sub>w<sub>i</sub></sub> +
&omega;<sub>W</sub>W<sub>w<sub>i</sub></sub> | w<sub>i</sub> &in; &Psi; }] /
 #&Psi;(&omega;<sub>V<sub>&nu;</sub></sub> +
&omega;<sub>I</sub> +
&omega;<sub>W</sub>)
</code>

<!--
TODO(CD): modify the colocation score to take priority into account
(e.g. preserving QoS for batch jobs shouldn't count for much compared to
high-priority production jobs.)
-->
