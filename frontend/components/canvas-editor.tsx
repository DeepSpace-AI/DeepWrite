"use client";

import { useEffect, useRef, useState, useCallback } from "react";
import {
  MousePointer2,
  Square,
  Circle,
  Type,
  Minus,
  Pencil,
  Eraser,
  Trash2,
  Undo,
  Redo,
  Download,
  ZoomIn,
  ZoomOut,
  Image as ImageIcon,
  Layers,
  Lock,
  Unlock,
  Eye,
  EyeOff,
} from "lucide-react";

interface CanvasEditorProps {
  width?: number;
  height?: number;
  initialImage?: string;
  onSave?: (dataUrl: string) => void;
  onExport?: (format: string, dataUrl: string) => void;
}

type Tool = "select" | "rect" | "circle" | "text" | "line" | "pencil" | "eraser" | "image";

interface LayerInfo {
  id: string;
  type: string;
  name: string;
  visible: boolean;
  locked: boolean;
}

export default function CanvasEditor({
  width = 800,
  height = 600,
  initialImage,
  onSave,
  onExport,
}: CanvasEditorProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const fabricRef = useRef<any>(null);
  const [activeTool, setActiveTool] = useState<Tool>("select");
  const [layers, setLayers] = useState<LayerInfo[]>([]);
  const [showLayers, setShowLayers] = useState(false);
  const [canUndo, setCanUndo] = useState(false);
  const [canRedo, setCanRedo] = useState(false);
  const [zoom, setZoom] = useState(1);
  const [color, setColor] = useState("#000000");
  const [fillColor, setFillColor] = useState("#transparent");
  const [strokeWidth, setStrokeWidth] = useState(2);
  const [fontSize, setFontSize] = useState(20);
  const historyRef = useRef<string[]>([]);
  const historyIndexRef = useRef(-1);

  const saveState = useCallback(() => {
    if (!fabricRef.current) return;
    const json = JSON.stringify(fabricRef.current.toJSON());
    historyRef.current = historyRef.current.slice(0, historyIndexRef.current + 1);
    historyRef.current.push(json);
    historyIndexRef.current = historyRef.current.length - 1;
    setCanUndo(historyIndexRef.current > 0);
    setCanRedo(false);
  }, []);

  const updateLayers = useCallback(() => {
    if (!fabricRef.current) return;
    const objects = fabricRef.current.getObjects();
    const newLayers: LayerInfo[] = objects.map((obj: any, index: number) => ({
      id: obj.id || `layer-${index}`,
      type: obj.type || "object",
      name: obj.name || `${obj.type} ${index + 1}`,
      visible: obj.visible !== false,
      locked: obj.selectable === false,
    }));
    setLayers(newLayers);
  }, []);

  useEffect(() => {
    if (!canvasRef.current || fabricRef.current) return;

    const loadFabric = async () => {
      const fabricModule = await import("fabric");
      const fabric = fabricModule;

      const canvas = new fabric.Canvas(canvasRef.current!, {
        width,
        height,
        backgroundColor: "#ffffff",
        selection: true,
      });

      fabricRef.current = canvas;

      if (initialImage) {
        const img = await fabric.FabricImage.fromURL(initialImage);
        img.scaleToWidth(width * 0.8);
        canvas.add(img);
        canvas.centerObject(img);
      }

      canvas.on("object:modified", saveState);
      canvas.on("object:added", () => {
        saveState();
        updateLayers();
      });
      canvas.on("object:removed", updateLayers);

      saveState();
      updateLayers();
    };

    loadFabric();

    return () => {
      if (fabricRef.current) {
        fabricRef.current.dispose();
        fabricRef.current = null;
      }
    };
  }, [width, height]);

  useEffect(() => {
    if (!fabricRef.current) return;
    const canvas = fabricRef.current;

    canvas.isDrawingMode = activeTool === "pencil" || activeTool === "eraser";

    if (activeTool === "pencil") {
      canvas.freeDrawingBrush = new (canvas.freeDrawingBrush as any).constructor(canvas, {
        color,
        width: strokeWidth,
      });
    } else if (activeTool === "eraser") {
      canvas.freeDrawingBrush = new (canvas.freeDrawingBrush as any).constructor(canvas, {
        color: "#ffffff",
        width: strokeWidth * 3,
      });
    }

    canvas.selection = activeTool === "select";
    canvas.defaultCursor = activeTool === "select" ? "default" : "crosshair";
  }, [activeTool, color, strokeWidth]);

  const handleToolClick = async (tool: Tool) => {
    if (!fabricRef.current) return;
    const canvas = fabricRef.current;

    setActiveTool(tool);

    if (tool === "rect") {
      const fabricModule = await import("fabric");
      const rect = new fabricModule.Rect({
        left: 100,
        top: 100,
        width: 150,
        height: 100,
        fill: fillColor === "transparent" ? "transparent" : fillColor,
        stroke: color,
        strokeWidth,
        name: `Rectangle ${layers.length + 1}`,
      });
      canvas.add(rect);
      canvas.setActiveObject(rect);
    } else if (tool === "circle") {
      const fabricModule = await import("fabric");
      const circle = new fabricModule.Circle({
        left: 200,
        top: 200,
        radius: 50,
        fill: fillColor === "transparent" ? "transparent" : fillColor,
        stroke: color,
        strokeWidth,
        name: `Circle ${layers.length + 1}`,
      });
      canvas.add(circle);
      canvas.setActiveObject(circle);
    } else if (tool === "text") {
      const fabricModule = await import("fabric");
      const text = new fabricModule.IText("Text", {
        left: 150,
        top: 150,
        fontSize,
        fill: color,
        name: `Text ${layers.length + 1}`,
      });
      canvas.add(text);
      canvas.setActiveObject(text);
    } else if (tool === "line") {
      const fabricModule = await import("fabric");
      const line = new fabricModule.Line([100, 100, 300, 100], {
        stroke: color,
        strokeWidth,
        name: `Line ${layers.length + 1}`,
      });
      canvas.add(line);
      canvas.setActiveObject(line);
    }

    if (tool !== "select" && tool !== "pencil" && tool !== "eraser") {
      setActiveTool("select");
    }
  };

  const handleUndo = () => {
    if (!fabricRef.current || historyIndexRef.current <= 0) return;
    historyIndexRef.current--;
    fabricRef.current.loadFromJSON(historyRef.current[historyIndexRef.current], () => {
      fabricRef.current?.renderAll();
      updateLayers();
    });
    setCanUndo(historyIndexRef.current > 0);
    setCanRedo(historyIndexRef.current < historyRef.current.length - 1);
  };

  const handleRedo = () => {
    if (!fabricRef.current || historyIndexRef.current >= historyRef.current.length - 1) return;
    historyIndexRef.current++;
    fabricRef.current.loadFromJSON(historyRef.current[historyIndexRef.current], () => {
      fabricRef.current?.renderAll();
      updateLayers();
    });
    setCanUndo(historyIndexRef.current > 0);
    setCanRedo(historyIndexRef.current < historyRef.current.length - 1);
  };

  const handleDelete = () => {
    if (!fabricRef.current) return;
    const activeObjects = fabricRef.current.getActiveObjects();
    activeObjects.forEach((obj: any) => fabricRef.current.remove(obj));
    fabricRef.current.discardActiveObject();
    fabricRef.current.renderAll();
  };

  const handleZoomIn = () => {
    if (!fabricRef.current) return;
    const newZoom = Math.min(zoom * 1.2, 3);
    fabricRef.current.setZoom(newZoom);
    setZoom(newZoom);
  };

  const handleZoomOut = () => {
    if (!fabricRef.current) return;
    const newZoom = Math.max(zoom / 1.2, 0.3);
    fabricRef.current.setZoom(newZoom);
    setZoom(newZoom);
  };

  const handleExport = (format: string) => {
    if (!fabricRef.current) return;
    let dataUrl: string;
    if (format === "svg") {
      dataUrl = fabricRef.current.toSVG();
    } else {
      dataUrl = fabricRef.current.toDataURL({
        format: format === "jpg" ? "jpeg" : format,
        quality: 0.9,
      });
    }
    onExport?.(format, dataUrl);

    const link = document.createElement("a");
    link.download = `canvas-export.${format}`;
    link.href = dataUrl;
    link.click();
  };

  const handleSave = () => {
    if (!fabricRef.current) return;
    const dataUrl = fabricRef.current.toDataURL({ format: "png", quality: 0.9 });
    onSave?.(dataUrl);
  };

  const toggleLayerVisibility = (index: number) => {
    if (!fabricRef.current) return;
    const obj = fabricRef.current.getObjects()[index];
    if (obj) {
      obj.visible = !obj.visible;
      fabricRef.current.renderAll();
      updateLayers();
    }
  };

  const toggleLayerLock = (index: number) => {
    if (!fabricRef.current) return;
    const obj = fabricRef.current.getObjects()[index];
    if (obj) {
      obj.selectable = !obj.selectable;
      obj.evented = obj.selectable;
      fabricRef.current.renderAll();
      updateLayers();
    }
  };

  const tools = [
    { id: "select" as Tool, icon: MousePointer2, label: "Select" },
    { id: "rect" as Tool, icon: Square, label: "Rectangle" },
    { id: "circle" as Tool, icon: Circle, label: "Circle" },
    { id: "line" as Tool, icon: Minus, label: "Line" },
    { id: "text" as Tool, icon: Type, label: "Text" },
    { id: "pencil" as Tool, icon: Pencil, label: "Pencil" },
    { id: "eraser" as Tool, icon: Eraser, label: "Eraser" },
  ];

  return (
    <div className="flex flex-col h-full bg-gray-100 rounded-lg overflow-hidden">
      <div className="flex items-center gap-1 px-3 py-2 bg-white border-b border-gray-200">
        <div className="flex items-center gap-1 border-r border-gray-200 pr-2 mr-2">
          {tools.map((tool) => (
            <button
              key={tool.id}
              onClick={() => handleToolClick(tool.id)}
              className={`p-2 rounded-lg transition-colors ${
                activeTool === tool.id
                  ? "bg-blue-100 text-blue-700"
                  : "hover:bg-gray-100 text-gray-600"
              }`}
              title={tool.label}
            >
              <tool.icon className="w-4 h-4" />
            </button>
          ))}
        </div>

        <div className="flex items-center gap-2 border-r border-gray-200 pr-2 mr-2">
          <input
            type="color"
            value={color}
            onChange={(e) => setColor(e.target.value)}
            className="w-8 h-8 rounded cursor-pointer"
            title="Stroke Color"
          />
          <input
            type="color"
            value={fillColor === "transparent" ? "#ffffff" : fillColor}
            onChange={(e) => setFillColor(e.target.value)}
            className="w-8 h-8 rounded cursor-pointer"
            title="Fill Color"
          />
          <select
            value={strokeWidth}
            onChange={(e) => setStrokeWidth(Number(e.target.value))}
            className="px-2 py-1 text-sm border rounded"
          >
            {[1, 2, 3, 5, 8, 10].map((w) => (
              <option key={w} value={w}>
                {w}px
              </option>
            ))}
          </select>
          <select
            value={fontSize}
            onChange={(e) => setFontSize(Number(e.target.value))}
            className="px-2 py-1 text-sm border rounded"
          >
            {[12, 16, 20, 24, 32, 48, 64].map((s) => (
              <option key={s} value={s}>
                {s}px
              </option>
            ))}
          </select>
        </div>

        <div className="flex items-center gap-1 border-r border-gray-200 pr-2 mr-2">
          <button
            onClick={handleUndo}
            disabled={!canUndo}
            className="p-2 rounded-lg hover:bg-gray-100 disabled:opacity-50"
            title="Undo"
          >
            <Undo className="w-4 h-4" />
          </button>
          <button
            onClick={handleRedo}
            disabled={!canRedo}
            className="p-2 rounded-lg hover:bg-gray-100 disabled:opacity-50"
            title="Redo"
          >
            <Redo className="w-4 h-4" />
          </button>
          <button
            onClick={handleDelete}
            className="p-2 rounded-lg hover:bg-red-100 text-red-600"
            title="Delete"
          >
            <Trash2 className="w-4 h-4" />
          </button>
        </div>

        <div className="flex items-center gap-1 border-r border-gray-200 pr-2 mr-2">
          <button
            onClick={handleZoomOut}
            className="p-2 rounded-lg hover:bg-gray-100"
            title="Zoom Out"
          >
            <ZoomOut className="w-4 h-4" />
          </button>
          <span className="text-sm text-gray-600 w-12 text-center">
            {Math.round(zoom * 100)}%
          </span>
          <button
            onClick={handleZoomIn}
            className="p-2 rounded-lg hover:bg-gray-100"
            title="Zoom In"
          >
            <ZoomIn className="w-4 h-4" />
          </button>
        </div>

        <div className="flex items-center gap-1">
          <button
            onClick={() => setShowLayers(!showLayers)}
            className={`p-2 rounded-lg ${showLayers ? "bg-blue-100 text-blue-700" : "hover:bg-gray-100"}`}
            title="Layers"
          >
            <Layers className="w-4 h-4" />
          </button>
          <button
            onClick={handleSave}
            className="px-3 py-1.5 bg-blue-600 text-white text-sm rounded-lg hover:bg-blue-700"
          >
            Save
          </button>
          <div className="relative group">
            <button className="px-3 py-1.5 bg-green-600 text-white text-sm rounded-lg hover:bg-green-700 flex items-center gap-1">
              <Download className="w-4 h-4" />
              Export
            </button>
            <div className="absolute right-0 top-full mt-1 bg-white rounded-lg shadow-lg border opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all z-10">
              {["png", "jpg", "svg", "pdf"].map((fmt) => (
                <button
                  key={fmt}
                  onClick={() => handleExport(fmt)}
                  className="block w-full px-4 py-2 text-sm text-left hover:bg-gray-100 first:rounded-t-lg last:rounded-b-lg"
                >
                  Export as {fmt.toUpperCase()}
                </button>
              ))}
            </div>
          </div>
        </div>
      </div>

      <div className="flex flex-1 overflow-hidden">
        <div className="flex-1 overflow-auto p-4 flex items-center justify-center">
          <div className="bg-white shadow-lg rounded-lg overflow-hidden">
            <canvas ref={canvasRef} />
          </div>
        </div>

        {showLayers && (
          <div className="w-64 bg-white border-l border-gray-200 overflow-y-auto">
            <div className="p-3 border-b border-gray-200">
              <h3 className="font-semibold text-sm">Layers</h3>
            </div>
            <div className="p-2">
              {layers.length === 0 ? (
                <p className="text-sm text-gray-500 text-center py-4">No layers</p>
              ) : (
                <div className="space-y-1">
                  {layers.map((layer, index) => (
                    <div
                      key={layer.id}
                      className="flex items-center gap-2 p-2 rounded hover:bg-gray-50"
                    >
                      <button
                        onClick={() => toggleLayerVisibility(index)}
                        className="p-1 hover:bg-gray-200 rounded"
                      >
                        {layer.visible ? (
                          <Eye className="w-3 h-3" />
                        ) : (
                          <EyeOff className="w-3 h-3 text-gray-400" />
                        )}
                      </button>
                      <button
                        onClick={() => toggleLayerLock(index)}
                        className="p-1 hover:bg-gray-200 rounded"
                      >
                        {layer.locked ? (
                          <Lock className="w-3 h-3 text-orange-500" />
                        ) : (
                          <Unlock className="w-3 h-3" />
                        )}
                      </button>
                      <span className="text-sm flex-1 truncate">{layer.name}</span>
                      <span className="text-xs text-gray-400">{layer.type}</span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
