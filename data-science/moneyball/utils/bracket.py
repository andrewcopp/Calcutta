import math


def sigmoid(x: float) -> float:
    if x >= 0:
        z = math.exp(-x)
        return 1.0 / (1.0 + z)
    z = math.exp(x)
    return z / (1.0 + z)


def win_prob(
    net1: float,
    net2: float,
    scale: float,
) -> float:
    if scale <= 0:
        raise ValueError("kenpom_scale must be positive")
    return sigmoid((net1 - net2) / scale)
