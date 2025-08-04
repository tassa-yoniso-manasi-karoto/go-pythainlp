#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
PyThaiNLP HTTP Service using aiohttp
Provides RESTful API for Go client
"""

import json
import time
import sys
import traceback
from aiohttp import web
import asyncio
from typing import Dict, List, Any, Optional

# Pre-load PyThaiNLP modules at startup
print("Loading PyThaiNLP modules...", file=sys.stderr)
start_time = time.time()

try:
    from pythainlp.tokenize import word_tokenize
    from pythainlp.transliterate import romanize, transliterate, pronunciate
    from pythainlp import __version__ as pythainlp_version
    
    # Pre-load engines to warm up
    _ = word_tokenize("ทดสอบ", engine="newmm")
    _ = romanize("ทดสอบ")
    
    print(f"PyThaiNLP {pythainlp_version} loaded in {time.time() - start_time:.2f}s", file=sys.stderr)
except Exception as e:
    print(f"Failed to load PyThaiNLP: {e}", file=sys.stderr)
    sys.exit(1)

# Dynamically detect available engines
def detect_available_engines():
    """Detect which engines are actually available based on installed dependencies"""
    tokenize_engines = []
    romanize_engines = []
    transliterate_engines = []
    
    # Always available tokenizers (dictionary-based)
    tokenize_engines.extend(["newmm", "longest", "nercut", "tltk"])
    
    # Check for ICU (requires pyicu)
    try:
        import icu
        tokenize_engines.append("icu")
    except ImportError:
        pass
    
    # Check for nlpo3 (Rust-based)
    try:
        import nlpo3
        tokenize_engines.append("nlpo3")
    except ImportError:
        pass
    
    # Check for neural tokenizers (require torch)
    try:
        import torch
        tokenize_engines.extend(["attacut", "deepcut", "oskut", "sefr_cut"])
    except ImportError:
        print("PyTorch not available - neural tokenizers disabled", file=sys.stderr)
    
    # Romanization engines
    romanize_engines.extend(["royin", "tltk", "lookup"])  # Always available
    
    # Check for thai2rom (requires torch)
    try:
        import torch
        romanize_engines.append("thai2rom")
    except ImportError:
        pass
    
    # Check for thai2rom_onnx
    try:
        import onnxruntime
        romanize_engines.append("thai2rom_onnx")
    except ImportError:
        pass
    
    # Transliteration engines
    transliterate_engines.extend(["iso_11940", "tltk_ipa", "tltk_g2p"])  # Always available
    
    # Check for ICU transliteration (requires pyicu)
    try:
        import icu
        transliterate_engines.append("icu")
    except ImportError:
        pass
    
    # Check for thaig2p and ipa (require torch/epitran)
    try:
        import torch
        transliterate_engines.extend(["thaig2p", "thaig2p_v2"])
    except ImportError:
        pass
    
    try:
        import epitran
        transliterate_engines.append("ipa")
    except ImportError:
        pass
    
    return tokenize_engines, romanize_engines, transliterate_engines

# Detect available engines at startup
TOKENIZE_ENGINES, ROMANIZE_ENGINES, TRANSLITERATE_ENGINES = detect_available_engines()
print(f"Available tokenizers: {TOKENIZE_ENGINES}", file=sys.stderr)
print(f"Available romanizers: {ROMANIZE_ENGINES}", file=sys.stderr)
print(f"Available transliterators: {TRANSLITERATE_ENGINES}", file=sys.stderr)


async def handle_tokenize(request: web.Request) -> web.Response:
    """Handle tokenization requests"""
    try:
        data = await request.json()
        text = data.get("text", "")
        engine = data.get("engine", "newmm")
        options = data.get("options", {})
        
        if not text:
            return web.json_response({
                "data": None,
                "metadata": {},
                "error": {
                    "code": "EMPTY_TEXT",
                    "message": "Text parameter is required"
                }
            }, status=400)
        
        if engine not in TOKENIZE_ENGINES:
            return web.json_response({
                "data": None,
                "metadata": {},
                "error": {
                    "code": "INVALID_ENGINE",
                    "message": f"Engine '{engine}' not supported",
                    "details": {"supported_engines": TOKENIZE_ENGINES}
                }
            }, status=400)
        
        start = time.time()
        tokens = word_tokenize(text, engine=engine, **options)
        processing_time = (time.time() - start) * 1000
        
        return web.json_response({
            "data": {
                "tokens": tokens
            },
            "metadata": {
                "engine": engine,
                "version": pythainlp_version,
                "processing_time_ms": round(processing_time, 2)
            },
            "error": None
        })
        
    except Exception as e:
        return web.json_response({
            "data": None,
            "metadata": {},
            "error": {
                "code": "INTERNAL_ERROR",
                "message": str(e),
                "details": {"traceback": traceback.format_exc()}
            }
        }, status=500)


async def handle_romanize(request: web.Request) -> web.Response:
    """Handle romanization requests"""
    try:
        data = await request.json()
        text = data.get("text", "")
        engine = data.get("engine", "royin")
        
        if not text:
            return web.json_response({
                "data": None,
                "metadata": {},
                "error": {
                    "code": "EMPTY_TEXT",
                    "message": "Text parameter is required"
                }
            }, status=400)
        
        if engine not in ROMANIZE_ENGINES:
            return web.json_response({
                "data": None,
                "metadata": {},
                "error": {
                    "code": "INVALID_ENGINE",
                    "message": f"Engine '{engine}' not supported",
                    "details": {"supported_engines": ROMANIZE_ENGINES}
                }
            }, status=400)
        
        start = time.time()
        
        # Tokenize first if requested
        if data.get("tokenize", False):
            tokens = word_tokenize(text)
            romanized_tokens = [romanize(token, engine=engine) for token in tokens]
            romanized_text = " ".join(romanized_tokens)
            result = {
                "romanized": romanized_text,
                "tokens": tokens,
                "romanized_tokens": romanized_tokens
            }
        else:
            romanized_text = romanize(text, engine=engine)
            result = {"romanized": romanized_text}
        
        processing_time = (time.time() - start) * 1000
        
        return web.json_response({
            "data": result,
            "metadata": {
                "engine": engine,
                "version": pythainlp_version,
                "processing_time_ms": round(processing_time, 2)
            },
            "error": None
        })
        
    except Exception as e:
        return web.json_response({
            "data": None,
            "metadata": {},
            "error": {
                "code": "INTERNAL_ERROR",
                "message": str(e),
                "details": {"traceback": traceback.format_exc()}
            }
        }, status=500)


async def handle_transliterate(request: web.Request) -> web.Response:
    """Handle transliteration (phonetic) requests"""
    try:
        data = await request.json()
        text = data.get("text", "")
        engine = data.get("engine", "thaig2p")
        
        if not text:
            return web.json_response({
                "data": None,
                "metadata": {},
                "error": {
                    "code": "EMPTY_TEXT",
                    "message": "Text parameter is required"
                }
            }, status=400)
        
        if engine not in TRANSLITERATE_ENGINES:
            return web.json_response({
                "data": None,
                "metadata": {},
                "error": {
                    "code": "INVALID_ENGINE",
                    "message": f"Engine '{engine}' not supported",
                    "details": {"supported_engines": TRANSLITERATE_ENGINES}
                }
            }, status=400)
        
        start = time.time()
        phonetic = transliterate(text, engine=engine)
        processing_time = (time.time() - start) * 1000
        
        return web.json_response({
            "data": {
                "phonetic": phonetic
            },
            "metadata": {
                "engine": engine,
                "version": pythainlp_version,
                "processing_time_ms": round(processing_time, 2)
            },
            "error": None
        })
        
    except Exception as e:
        return web.json_response({
            "data": None,
            "metadata": {},
            "error": {
                "code": "INTERNAL_ERROR",
                "message": str(e),
                "details": {"traceback": traceback.format_exc()}
            }
        }, status=500)


async def handle_analyze(request: web.Request) -> web.Response:
    """Handle combined analysis requests"""
    try:
        data = await request.json()
        text = data.get("text", "")
        features = data.get("features", ["tokenize", "romanize"])
        
        if not text:
            return web.json_response({
                "data": None,
                "metadata": {},
                "error": {
                    "code": "EMPTY_TEXT",
                    "message": "Text parameter is required"
                }
            }, status=400)
        
        start = time.time()
        result = {}
        
        # Always tokenize first as base
        tokens = word_tokenize(text, engine=data.get("tokenize_engine", "newmm"))
        if "tokenize" in features:
            result["tokens"] = tokens
        
        if "romanize" in features:
            engine = data.get("romanize_engine", "royin")
            romanized_tokens = [romanize(token, engine=engine) for token in tokens]
            result["romanized"] = " ".join(romanized_tokens)
            result["romanized_tokens"] = romanized_tokens
        
        if "transliterate" in features:
            engine = data.get("transliterate_engine", "thaig2p")
            result["phonetic"] = transliterate(text, engine=engine)
        
        processing_time = (time.time() - start) * 1000
        
        return web.json_response({
            "data": result,
            "metadata": {
                "features": features,
                "version": pythainlp_version,
                "processing_time_ms": round(processing_time, 2)
            },
            "error": None
        })
        
    except Exception as e:
        return web.json_response({
            "data": None,
            "metadata": {},
            "error": {
                "code": "INTERNAL_ERROR",
                "message": str(e),
                "details": {"traceback": traceback.format_exc()}
            }
        }, status=500)


async def handle_health(request: web.Request) -> web.Response:
    """Health check endpoint"""
    return web.json_response({
        "status": "ready",
        "version": pythainlp_version,
        "engines": {
            "tokenize": TOKENIZE_ENGINES,
            "romanize": ROMANIZE_ENGINES,
            "transliterate": TRANSLITERATE_ENGINES
        }
    })


def create_app() -> web.Application:
    """Create and configure the web application"""
    app = web.Application()
    
    # Add routes
    app.router.add_post('/tokenize', handle_tokenize)
    app.router.add_post('/romanize', handle_romanize)
    app.router.add_post('/transliterate', handle_transliterate)
    app.router.add_post('/analyze', handle_analyze)
    app.router.add_get('/health', handle_health)
    
    return app


if __name__ == '__main__':
    app = create_app()
    print("Starting PyThaiNLP HTTP service on port __PYTHAINLP_SERVICE_PORT__...", file=sys.stderr)
    web.run_app(app, host='0.0.0.0', port=__PYTHAINLP_SERVICE_PORT__)