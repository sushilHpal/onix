import * as React from "react";
import { useEffect } from "react";
import "@patternfly/react-core/dist/styles/base.css";
import { BrowserRouter as Router, Redirect, Switch, withRouter } from "react-router-dom";
import axios from "axios";
import { useDispatch, useSelector } from "react-redux";

import ConsoleLayout from "./ConsoleLayout/ConsoleLayout";
import ConsoleRoutes, { IAppRoute } from "./ConsoleLayout/ConsoleRoutes";
import MetaModel from "./MetaModel";
import { Dashboard } from "./Dashboard/Dashboard";
import { ACTIONS } from "./MetaModel/data/metamodelReducer";

const Console: React.ComponentClass<{}> = withRouter((props) => {

  const dispatch = useDispatch();
  const dashboardRoutes: IAppRoute[] = [
    {
      component: Dashboard,
      exact: true,
      label: "Dashboard",
      path: "dashboard",
      title: "Onix Dashboard"
    }
  ];
  const metaModelRoutes: IAppRoute[] = useSelector(store => {
      const results: IAppRoute[] = [];
      store.MetaModelReducer.models.forEach((item, index) => {
          results.push(
            {
              component: MetaModel,
              exact: true,
              label: item.name,
              path: `metamodel`,
              title: item.name
            }
          );
        }
      );
      return results;
    }
  );

  useEffect(() => {
    axios.get("/api/model")
      .then((response) => {
          dispatch({type: ACTIONS.SET_MODELS, models: response.data.values});
        }
      ).catch((error) => {
      console.log(error);
    });
  }, []);

  return (
    <Router>
      <ConsoleLayout dashboardRoutes={dashboardRoutes} metaModelRoutes={metaModelRoutes}>
        <Switch>
          <ConsoleRoutes routes={[...dashboardRoutes, ...metaModelRoutes]} baseurl={props.match.url}/>
        </Switch>
      </ConsoleLayout>
      <Redirect to={"/console/dashboard"}/>
    </Router>);
});

export default Console;
