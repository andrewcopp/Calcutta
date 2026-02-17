import math


def round_order(round_name: str) -> int:
    order = {
        "first_four": 1,
        "round_of_64": 2,
        "round_of_32": 3,
        "sweet_16": 4,
        "elite_8": 5,
        "final_four": 6,
        "championship": 7,
    }
    return int(order.get(str(round_name), 999))


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
