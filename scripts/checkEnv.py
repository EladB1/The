from typing import List, Dict
import os
import json

def prepareEnvironmentVariables(keys: List[str]) -> Dict[str, str]:
    values = dict()
    for key in keys:
        values[key] = os.environ.get(key)
    return values


if __name__ == '__main__':
    vars = prepareEnvironmentVariables([
        'THE_DEV_DEBUG',
        'THE_DEV_LITERALS',
        'THE_DEV_LEXER',
        'THE_DEV_PARSER',
        'THE_DEV_SCOPES',
        'THE_DEV_ANNOTATED',
        'THE_DEV_IRGEN',
        'THE_DEV_CODEGEN'
    ])
    print(json.dumps(vars))