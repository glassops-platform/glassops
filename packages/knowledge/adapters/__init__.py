# generation/adapters/__init__.py
"""Language adapters for documentation generation."""

from .base import BaseAdapter
from .go import GoAdapter
from .python import PythonAdapter
from .typescript import TypeScriptAdapter
from .yaml_adapter import YAMLAdapter
from .json_adapter import JSONAdapter
from .docker_adapter import DockerAdapter
from .terraform_adapter import TerraformAdapter
from .apex_adapter import ApexAdapter
from .lwc_adapter import LWCAdapter

__all__ = [
    "BaseAdapter",
    "GoAdapter",
    "PythonAdapter",
    "TypeScriptAdapter",
    "YAMLAdapter",
    "JSONAdapter",
    "DockerAdapter",
    "TerraformAdapter",
    "ApexAdapter",
    "LWCAdapter",
]
