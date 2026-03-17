"""Tests for hello module."""

from src.hello import add, greet


def test_greet():
    assert greet("World") == "Hello, World!"


def test_add():
    assert add(1, 2) == 3
