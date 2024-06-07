// MIT License
//
// Copyright (c) 2024 Marcel Joachim Kloubert (https://marcel.coffee)
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// page component
const GoDependencyGraphPage = () => {
  const [selectedModule, setSelectedModule] = React.useState();
  const [scale, setScale] = React.useState('100');

  const graphContainerRef = React.useRef(null);

  const {
    appName,
    graphDirection,
    graphHeight,
    graphScalePercentage,
    graphWidth,
    infoboxWidth,
    mermaidGraph,
    moduleList,
    sidebarWidth,
  } = useGlobalVars();

  const scaleFactor = React.useMemo(() => {
    const scalePercentage = parseFloat(scale.trim());
    if (!Number.isNaN(scalePercentage)) {
      return scalePercentage / 100.0;
    }

    return 1.0;
  }, [scale]);

  const contentWidth = React.useMemo(() => {
    return `calc(100% - ${sidebarWidth})`;
  }, [sidebarWidth]);

  const finalGraphCode = React.useMemo(() => {
    return mermaidGraph.split('<<<GraphDirection>>>')
      .join(graphDirection);
  }, [mermaidGraph]);

  const renderModuleList = React.useCallback(() => {
    return (
      moduleList.map((m, mIndex) => {
        return (
          <li
            key={`project-module-${mIndex}`}
            style={{
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              width: `calc(${sidebarWidth} - 40px)`,
              wordBreak: 'break-all',
              cursor: 'pointer',
              fontSize: '0.8rem',
            }}
            onClick={() => {
              setSelectedModule(m);
            }}
          >{m.Name}@{m.Version}</li>
        );
      })
    );
  }, [moduleList]);

  const renderInfoBox = React.useCallback(() => {
    if (!selectedModule) {
      return (
        <div className="h-full w-full flex items-center justify-center">
          <p className="text-center">
            Click on a module of the left side to see more information here ...
          </p>
        </div>
      );
    }

    return (
      <div
        className="h-full w-full px-6 py-4 bg-gray-300"
      >
        <div
          className="font-bold text-sm mb-2 w-full"
          style={{
            wordBreak: 'break-word'
          }}
        >
          {selectedModule.Name}
        </div>

        {!!selectedModule.Link && (
          <a
            className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded inline-block text-center w-full block"
            href={selectedModule.Link} target="_blank"
          >
            Open
          </a>
        )}
      </div>
    );
  }, [selectedModule]);

  React.useEffect(() => {
    mermaid.initialize({
      startOnLoad: false,
      flowchart: {
        useMaxWidth: true,
        htmlLabels: true,
      }
    });

    setScale(graphScalePercentage.toFixed(2));
  }, []);

  React.useEffect(() => {
    const el = graphContainerRef.current;
    if (!el) {
      return;
    }

    el.innerHTML = "";
    mermaid.render('gpmDependencyGraph', finalGraphCode).then(({ svg }) => {
      el.innerHTML = svg;
    });
  }, [finalGraphCode, graphDirection]);

  return (
    <React.Fragment>
      {/** left navbar */}
      <div
        className="bg-gray-800 text-white w-64 h-full p-4 overflow-y-auto overflow-x-hidden"
        style={{
          width: sidebarWidth,
        }}
      >
        <h1 className="text-xl font-bold mb-4">{appName}</h1>

        {/** scale input field */}
        <div className="mb-4">
          <label htmlFor="gpm-graph-scale" className="block mb-2">Scale (%)</label>
          <input
            type="number"
            id="gpm-graph-scale"
            className="text-gray-900 p-2 w-full"
            value={scale}
            onChange={(e) => {
              setScale(e.target.value);
            }}
          />
        </div>

        {/** module list */}
        <div className="mb-4">
          <label htmlFor="gpm-module-list" className="block mb-2">Installed Modules</label>
          <div
            id="gpm-module-list"
            className="mb-4 moduleList"
            style={{
              width: `calc(${sidebarWidth} - 40px)`,
            }}
          >
            {renderModuleList()}
          </div>
        </div>
      </div>

      {/** middle content */}
      <div
        id="gpm-dependency-graph-container"
        className="bg-white text-gray-800 h-full p-4 overflow-auto"
        style={{
          width: contentWidth,
        }}
      >
        {/** final Mermaid graph */}
        <div
          ref={graphContainerRef}
          className="mermaid"
          style={{
            display: 'block',
            height: `calc(${graphHeight} * ${scaleFactor})`,
            maxHeight: `calc(${graphHeight} * ${scaleFactor})`,
            maxWidth: `calc(${graphWidth} * ${scaleFactor})`,
            width: `calc(${graphWidth} * ${scaleFactor})`,
          }}
        />
      </div>

      {/** right infobox */}
      <div
        className="max-w-sm rounded overflow-hidden shadow-lg bg-gray-100"
        id="gpm-module-info-box"
        style={{
          width: infoboxWidth,
        }}
      >
        {renderInfoBox()}
      </div>
    </React.Fragment>
  );
};

// render into #gpm-content
ReactDOM.createRoot(document.querySelector("#gpm-content")).render(
  <GoDependencyGraphPage />
);
