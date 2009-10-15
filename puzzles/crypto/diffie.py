prime = 51237129793
mod = 321454621
a = 341
A = prime ** a % mod
b = 573
B = prime ** b % mod
key = A**b % mod
assert B**a % mod == key, 'Bad diffie math.'
