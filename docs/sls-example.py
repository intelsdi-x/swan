#!/usr/bin/env python

import numpy as np
import matplotlib.pyplot as plt

num_samples = 100
jitter = 0.2

def smooth(y, box_pts):
    box = np.ones(box_pts) / box_pts
    y_smooth = np.convolve(y, box, mode='same')
    return y_smooth

x = np.linspace(0.00000001, np.pi, num_samples)
random_field = (np.random.ranf(num_samples) - 0.5) * jitter
smooth_random_field = smooth(random_field, 3)
sls = smooth_random_field + (1 + 0.15 * np.sin(x + np.pi/2))
slo = np.linspace(1, 1, num_samples)

plt.plot(x, slo, label='SLO', color='blue')
plt.plot(x, sls, label='SLS', color='black')
plt.plot(x, x * 0)

plt.fill_between(x, slo, sls, where=sls<1.0, interpolate=True, color='red')
plt.fill_between(x, slo, sls, where=sls>1.0, interpolate=True, color='green')

plt.title("Workload Performance")
plt.xlabel('Time')
plt.ylabel('Service Level Score')
plt.legend()

plt.savefig('sls-example.png')
