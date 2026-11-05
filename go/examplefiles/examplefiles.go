package examplefiles

// SystemParameters is netlogo/_SYSTEM_Parameters.csv, complete (2 lines).
const SystemParameters = `AbundanceBoost,BeeSpeciesInitialQueensListAsString,Backgroundcolour,BackgroundImage,Black_th,COLONIES_IBM,CropRotationListAsString,CumulVisitsOnlyLastYear?,FlowerspeciesFile ,FoodSourceLimit,ForagingMortalityFactor,ForagingMortalityModel,Gridsize,HabitatsFile,InitialQueensModifierBehaviorSpace,KeepDeadColonies?,Lambda_detectProb,MapAreaIncluded,MasterSizeFactor,MaxForagingRange_m,MaxHibernatingQueens,MaxPatchRadius_m,MergeHedges?,MinSizeFoodSources?,MyFarmlandPatches,N_Badgers,RemoveEmptyFoodSources?,SatelliteFile,SetColour,SexLocus?,ShowCohorts?,ShowDeadCols?,ShowFoodsources?,ShowGrid?,ShowInspectedColony?,ShowMasterpatchesOnly?,ShowNests?,ShowPlots?,ShowQueens?,ShowSearchingQueens?,ShowWeather?,BeespeciesFile,StopExtinct?,UnlimitedMales?,Weather,White_th,WinterMortality?
1,"""B_terrestris 500""",5,"""none""",1,0,"""""",FALSE,"""_SYSTEM_Flowerspecies.csv""",25,1,"""high""",500,"""_SYSTEM_Habitats.csv""",1,TRUE,-0.005,"""complete""",1,10000,10000,500,TRUE,TRUE,"""""",0,TRUE,"""No satellite image""","""Yellow""",FALSE,TRUE,FALSE,TRUE,FALSE,FALSE,FALSE,TRUE,TRUE,TRUE,TRUE,TRUE,"""_SYSTEM_BumbleSpecies_UK_01.csv""",TRUE,TRUE,"""Constant 8 hrs""",9,TRUE
`

// ExampleFarmParameters is netlogo/_SYSTEM_Example_Farm_Parameters.csv, complete (2 lines).
// With the default interface settings NetLogo reads this file (finding 13, see trans/netlogo-findings.md#finding-13).
const ExampleFarmParameters = `AbundanceBoost,BeeSpeciesInitialQueensListAsString,Backgroundcolour,BackgroundImage,Black_th,COLONIES_IBM,CropRotationListAsString,CumulVisitsOnlyLastYear?,"FlowerspeciesFile ",FoodSourceLimit,ForagingMortalityFactor,ForagingMortalityModel,Gridsize,HabitatsFile,InitialQueensModifierBehaviorSpace,KeepDeadColonies?,Lambda_detectProb,MapAreaIncluded,MasterSizeFactor,MaxForagingRange_m,MaxHibernatingQueens,MaxPatchRadius_m,MergeHedges?,MinSizeFoodSources?,MyFarmlandPatches,N_Badgers,RemoveEmptyFoodSources?,SatelliteFile,SetColour,SexLocus?,ShowCohorts?,ShowDeadCols?,ShowFoodsources?,ShowGrid?,ShowInspectedColony?,ShowMasterpatchesOnly?,ShowNests?,ShowPlots?,ShowQueens?,ShowSearchingQueens?,ShowWeather?,BeespeciesFile,StopExtinct?,UnlimitedMales?,Weather,White_th,WinterMortality?
1,"""B_terrestris 500""",5,"""none""",1,0,"""""",false,"""_SYSTEM_Flowerspecies.csv""",25,1,"""high""",500,"""_SYSTEM_Habitats.csv""",1,true,-0.005,"""complete""",1,10000,10000,500,true,true,"""""",0,true,"""No satellite image""","""Magenta""",false,true,false,true,false,false,false,true,true,true,true,true,"""_SYSTEM_BumbleSpecies_UK_01.csv""",true,true,"""Constant 8 hrs""",9,true
`

// Habitats contains lines of netlogo/_SYSTEM_Habitats.csv: the header, the first 2 data lines (lines 2-3 of the
// file) and the 6 stewardship option lines at the end of the file (lines 27-32, habitatColourID "FlowerSpeciesList_...").
const Habitats = `habitatColourID,colourName_forUserInfoOnly,colourRangeMin,colourRangeMax,habitatType,habitatSwitchedOn?,ONLY FLOWER DATA BEYOND THIS COLUMN!,Alsike_clover,American_willowherb,Average_Willow,Bell_heather,Birdsfoot_trefoil,Blackthorn,Bramble,Buckwheat,Bugle,Burdock,Buttercup,Charlock,Chicory,Comfrey,Common_bluebell,Common_cats_ear,Common_field_speedwell,Common_fumitory,Common_knapweed,Common_mouse_ear,Common_vetch,Creeping_buttercup,Crop_Apple,Crop_Cereals,Crop_Field_beans,Crop_Maize,Crop_Oilseed_rape,Cross_leaved_heath,Daisy,Dandelion,Dog_rose,Field_pansy,Field_scabious,Field_woundwort,Foxglove,Germander_speedwell,Giant_bindweed,Gorse,Greater_birdsfoot_trefoil,Greater_knapweed,Ground_ivy,Hawthorn,Heath_speedwell,Hedge_bindweed,Hedge_woundwort,Ling,Marsh_thistle,Meadow_buttercup,Meadowsweet,Oilseed_rape,Ox-eye_daisy,Phacelia,Prickly_sow_thistle,Ragwort,Red_campion,Red_clover,Red_dead_nettle,Redshank,Ribwort_plantain,Rosebay_willowherb,Salad_burnet,Scarlet_pimpernel,Selfheal,Silverweed,Smooth_hawksbeard,Spear_thistle,St_Johns_wort,Sunflower,Tormentil,Tufted_vetch,Vipers_bugloss,White_clover,White_dead_nettle,White_mustard,Wild_carrot,Wild_radish,Wild_teasel,Yarrow,Yellow_rattle
12,darkred,12,12,Grass_margin,TRUE,n/a,0.0175,0,0,0,0.376666667,0,0,0,0,0,1.528333333,0,0.026666667,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0.003333333,0,0.01,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0.129333333,0,0,0,0,0,0.14,0,0,0.046666667,0.02,0,0,0,0,0,0.276,0,0,0.203333333,0.098333333,0,0,0
15,red,13,17,Crop_Field_beans,TRUE,n/a,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,107.5,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0
FlowerSpeciesList_Legume,,-1,-1,Legume,FALSE,n/a,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,17.02,0,0,0,0,0,0,0,0.1077,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1.123,0,0,0,0,0,0,0,0,0,0,0,0,16.13077,0,0,0,0,0,0,0
FlowerSpeciesList_Margin,,-1,-1,Margin,FALSE,n/a,0,0,0,0,0.236666667,0,0,0,0,0,0,0,6.154444444,0,0,0.011666667,0,0.333333333,0.103333333,0,0.016666667,0,0,0,0,0,0,0,0,0,0,0.102222222,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,3.186666667,0.024444444,0.082222222,0.041111111,2.615333333,0,0,0,0,0,0.12,0,0,0,0.003333333,0,0,0,0,0,1.9,0,0,0.568333333,1.823333333,0,0,0
FlowerSpeciesList_newCSO_Legume,,-1,-1,newCSO_Legume,FALSE,n/a,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,17.02,0,0,0,0,0,0,0,0.1077,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1.123,0,0,0,0,0,0,0,0,0,0,0,0,16.13077,0,0,0,0,0,0,0
FlowerSpeciesList_newCSO_Margin,,-1,-1,BirdBumble_Margin,FALSE,n/a,0.49,0.42,0,0,0.01,0,0,1.53,0,0,0,0,0,0,0,0,0.21,0.23,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0.21,0,0,0,0,0,0.15,0,1.86,0,0,0,0.02,0,0,0,0.01,0,0.02,0,0,0.07,0.73,0,0.01,0,0,0,0,0,1.84,0,0,0,0.01,0
FlowerSpeciesList_newCSO_Plot,,-1,-1,GameCover_Plot,FALSE,n/a,0,0,0,0,0,0,0,6.276666667,0,0,0,0,0,0,0,0,4.79,0,0,0,0.16,0,0,0,0,0,0,0,0,0,0,0,0,0.7,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,4.47,0.19,0,0,0,0.16,0,0,0,0,1.185,0,0,0,0,0,2.2,0,0,0,0,0,0.82,0,17.37333333,0,0,0
FlowerSpeciesList_Plot,,-1,-1,Plot,FALSE,n/a,0,0,0,0,5.68,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1.058333333,0,30.805,0,0,0,0,0,0,0,1.67E-02,0,0,0,0,0,0.16875,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,7.855,0.033,0,0,0,0,0,0.01875,0,0.10625,0,0,0,0,0,0,1.725,0,0,0,0,0,0,0
`

// ExampleFarmFoodsources is netlogo/_SYSTEM_Example_Farm_Foodsources.txt, complete (76 lines, LF line endings as the original).
// Line 1: metres per NetLogo patch, line 2: number of flower patches, line 3: header, then 73 flower patches.
const ExampleFarmFoodsources = `10
73
id patchType patchColour xcor ycor size_sqm flowerSpeciesList perimeter_m info
0 "Hedgerow" 105 130 175 200 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 20  "no info" 
1 "Crop_Cereals" 75 125.948 141.856 105900 [["Crop_Cereals" 1]] 2080  "no info" 
2 "Hedgerow" 105 132 170.5 400 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 40  "no info" 
3 "Hedgerow" 105 133 167 100 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 10  "no info" 
4 "Hedgerow" 105 133 162 500 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 50  "no info" 
5 "Hedgerow" 105 132 155.5 600 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 60  "no info" 
6 "Crop_Cereals" 75 96 124 85700 [["Crop_Cereals" 1]] 2400  "no info" 
7 "Hedgerow" 105 131 151 100 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 10  "no info" 
8 "Hedgerow" 105 107 121 11700 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 1170  "no info" 
9 "Hedgerow" 105 131.333 147 300 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 30  "no info" 
10 "Crop_Cereals" 75 158.111 130.564 104600 [["Crop_Cereals" 1]] 2900  "no info" 
11 "Hedgerow" 105 179 124 18000 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 1800  "no info" 
12 "Hedgerow" 105 129 132 3700 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 370  "no info" 
13 "Hedgerow" 105 181 141 100 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 10  "no info" 
14 "Hedgerow" 105 184 139 300 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 30  "no info" 
15 "Crop_Cereals" 75 196.446 115.052 28900 [["Crop_Cereals" 1]] 1080  "no info" 
16 "Scrub" 125 85.119 122.903 13400 [["Average_Willow" 1.208] ["Birdsfoot_trefoil" 0.215] ["Blackthorn" 1.292] ["Bramble" 0.211] ["Bugle" 0.509] ["Burdock" 0.05] ["Buttercup" 0.792] ["Common_bluebell" 0.256] ["Common_knapweed" 0.56] ["Common_vetch" 0.011] ["Dandelion" 0.113] ["Dog_rose" 0.055] ["Field_scabious" 0.023] ["Foxglove" 0.086] ["Giant_bindweed" 0.032] ["Greater_knapweed" 0.029] ["Ground_ivy" 0.095] ["Hawthorn" 0.601] ["Hedge_woundwort" 0.287] ["Marsh_thistle" 0.016] ["Oilseed_rape" 0.018] ["Ragwort" 0.063] ["Red_clover" 0.046] ["Red_dead_nettle" 0.028] ["Rosebay_willowherb" 0.043] ["Selfheal" 0.06] ["Spear_thistle" 0.006] ["St_Johns_wort" 0.011] ["Tufted_vetch" 0.026] ["Vipers_bugloss" 0.5] ["White_clover" 0.76] ["White_dead_nettle" 0.171] ["Wild_teasel" 0.037]] 500  "no info" 
17 "Hedgerow" 105 194 128 100 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 10  "no info" 
18 "Hedgerow" 105 203.578 107.078 11600 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 1160  "no info" 
19 "Grass_improved" 115 166.523 104.438 161300 [["Dandelion" 0.006] ["White_clover" 0.47]] 3210  "no info" 
20 "Scrub" 125 202 121.5 3200 [["Average_Willow" 1.208] ["Birdsfoot_trefoil" 0.215] ["Blackthorn" 1.292] ["Bramble" 0.211] ["Bugle" 0.509] ["Burdock" 0.05] ["Buttercup" 0.792] ["Common_bluebell" 0.256] ["Common_knapweed" 0.56] ["Common_vetch" 0.011] ["Dandelion" 0.113] ["Dog_rose" 0.055] ["Field_scabious" 0.023] ["Foxglove" 0.086] ["Giant_bindweed" 0.032] ["Greater_knapweed" 0.029] ["Ground_ivy" 0.095] ["Hawthorn" 0.601] ["Hedge_woundwort" 0.287] ["Marsh_thistle" 0.016] ["Oilseed_rape" 0.018] ["Ragwort" 0.063] ["Red_clover" 0.046] ["Red_dead_nettle" 0.028] ["Rosebay_willowherb" 0.043] ["Selfheal" 0.06] ["Spear_thistle" 0.006] ["St_Johns_wort" 0.011] ["Tufted_vetch" 0.026] ["Vipers_bugloss" 0.5] ["White_clover" 0.76] ["White_dead_nettle" 0.171] ["Wild_teasel" 0.037]] 220  "no info" 
21 "Woodland" 35 77.687 106.287 52700 [["Birdsfoot_trefoil" 0.036] ["Blackthorn" 0.333] ["Bramble" 0.419] ["Bugle" 0.298] ["Burdock" 0.104] ["Buttercup" 0.724] ["Common_bluebell" 0.599] ["Common_knapweed" 0.008] ["Common_vetch" 0.038] ["Dandelion" 0.01] ["Dog_rose" 0.042] ["Field_scabious" 0.004] ["Foxglove" 0.02] ["Ground_ivy" 0.102] ["Hawthorn" 2.733] ["Hedge_woundwort" 0.038] ["Marsh_thistle" 0.023] ["Ragwort" 0.003] ["Red_clover" 0.005] ["Red_dead_nettle" 0.025] ["Rosebay_willowherb" 0.02] ["Selfheal" 0.08] ["Spear_thistle" 0.007] ["St_Johns_wort" 0.01] ["Tufted_vetch" 0.002] ["White_clover" 0.083] ["White_dead_nettle" 0.008]] 1440  "no info" 
22 "Grassland" 135 111.035 113.783 11500 [["Birdsfoot_trefoil" 0.352] ["Blackthorn" 0.225] ["Bramble" 0.017] ["Bugle" 0.028] ["Burdock" 0.017] ["Buttercup" 2.161] ["Common_bluebell" 0.487] ["Common_knapweed" 0.011] ["Common_vetch" 0.034] ["Dandelion" 0.121] ["Field_scabious" 0.017] ["Ground_ivy" 0.059] ["Hawthorn" 0.617] ["Hedge_woundwort" 0.004] ["Ragwort" 0.023] ["Red_clover" 0.253] ["Red_dead_nettle" 0.026] ["Selfheal" 0.113] ["Spear_thistle" 0.005] ["Tufted_vetch" 0.117] ["White_clover" 1.667] ["White_dead_nettle" 0.025]] 630  "no info" 
23 "Crop_Cereals" 75 218.941 97.295 99000 [["Crop_Cereals" 1]] 2870  "no info" 
24 "Grass_improved" 115 133.548 112.5 4200 [["Dandelion" 0.006] ["White_clover" 0.47]] 180  "no info" 
25 "Hedgerow" 105 221.5 117 200 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 20  "no info" 
26 "Woodland" 35 125.491 100.888 21400 [["Birdsfoot_trefoil" 0.036] ["Blackthorn" 0.333] ["Bramble" 0.419] ["Bugle" 0.298] ["Burdock" 0.104] ["Buttercup" 0.724] ["Common_bluebell" 0.599] ["Common_knapweed" 0.008] ["Common_vetch" 0.038] ["Dandelion" 0.01] ["Dog_rose" 0.042] ["Field_scabious" 0.004] ["Foxglove" 0.02] ["Ground_ivy" 0.102] ["Hawthorn" 2.733] ["Hedge_woundwort" 0.038] ["Marsh_thistle" 0.023] ["Ragwort" 0.003] ["Red_clover" 0.005] ["Red_dead_nettle" 0.025] ["Rosebay_willowherb" 0.02] ["Selfheal" 0.08] ["Spear_thistle" 0.007] ["St_Johns_wort" 0.01] ["Tufted_vetch" 0.002] ["White_clover" 0.083] ["White_dead_nettle" 0.008]] 990  "no info" 
27 "Hedgerow" 105 238 109 300 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 30  "no info" 
28 "Hedgerow" 105 145 107 7400 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 740  "no info" 
29 "Hedgerow" 105 116 105.5 400 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 40  "no info" 
30 "Crop_Cereals" 75 128.85 83.948 156200 [["Crop_Cereals" 1]] 2500  "no info" 
31 "Crop_Cereals" 75 97.97 91.255 36500 [["Crop_Cereals" 1]] 940  "no info" 
32 "Hedgerow" 105 225 100 2900 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 290  "no info" 
33 "Hedgerow" 105 121.786 98 1400 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 140  "no info" 
34 "Grassland" 135 191.922 86.398 33200 [["Birdsfoot_trefoil" 0.352] ["Blackthorn" 0.225] ["Bramble" 0.017] ["Bugle" 0.028] ["Burdock" 0.017] ["Buttercup" 2.161] ["Common_bluebell" 0.487] ["Common_knapweed" 0.011] ["Common_vetch" 0.034] ["Dandelion" 0.121] ["Field_scabious" 0.017] ["Ground_ivy" 0.059] ["Hawthorn" 0.617] ["Hedge_woundwort" 0.004] ["Ragwort" 0.023] ["Red_clover" 0.253] ["Red_dead_nettle" 0.026] ["Selfheal" 0.113] ["Spear_thistle" 0.005] ["Tufted_vetch" 0.117] ["White_clover" 1.667] ["White_dead_nettle" 0.025]] 1300  "no info" 
35 "Hedgerow" 105 235 98 100 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 10  "no info" 
36 "Crop_Cereals" 75 124.333 95.333 300 [["Crop_Cereals" 1]] 30  "no info" 
37 "Hedgerow" 105 208.5 96 200 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 20  "no info" 
38 "Hedgerow" 105 167.643 88.5 1400 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 140  "no info" 
39 "Hedgerow" 105 234 94 300 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 30  "no info" 
40 "Hedgerow" 105 107.879 90.606 3300 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 330  "no info" 
41 "Hedgerow" 105 184 86 1700 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 170  "no info" 
42 "Hedgerow" 105 192.4 92.4 500 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 50  "no info" 
43 "Hedgerow" 105 209 89 1200 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 120  "no info" 
44 "Hedgerow" 105 233 90 100 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 10  "no info" 
45 "Hedgerow" 105 232 88 100 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 10  "no info" 
46 "Woodland" 35 219 86 100 [["Birdsfoot_trefoil" 0.036] ["Blackthorn" 0.333] ["Bramble" 0.419] ["Bugle" 0.298] ["Burdock" 0.104] ["Buttercup" 0.724] ["Common_bluebell" 0.599] ["Common_knapweed" 0.008] ["Common_vetch" 0.038] ["Dandelion" 0.01] ["Dog_rose" 0.042] ["Field_scabious" 0.004] ["Foxglove" 0.02] ["Ground_ivy" 0.102] ["Hawthorn" 2.733] ["Hedge_woundwort" 0.038] ["Marsh_thistle" 0.023] ["Ragwort" 0.003] ["Red_clover" 0.005] ["Red_dead_nettle" 0.025] ["Rosebay_willowherb" 0.02] ["Selfheal" 0.08] ["Spear_thistle" 0.007] ["St_Johns_wort" 0.01] ["Tufted_vetch" 0.002] ["White_clover" 0.083] ["White_dead_nettle" 0.008]] 10  "no info" 
47 "Hedgerow" 105 221 70 6500 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 650  "no info" 
48 "Woodland" 35 223.425 65.636 34100 [["Birdsfoot_trefoil" 0.036] ["Blackthorn" 0.333] ["Bramble" 0.419] ["Bugle" 0.298] ["Burdock" 0.104] ["Buttercup" 0.724] ["Common_bluebell" 0.599] ["Common_knapweed" 0.008] ["Common_vetch" 0.038] ["Dandelion" 0.01] ["Dog_rose" 0.042] ["Field_scabious" 0.004] ["Foxglove" 0.02] ["Ground_ivy" 0.102] ["Hawthorn" 2.733] ["Hedge_woundwort" 0.038] ["Marsh_thistle" 0.023] ["Ragwort" 0.003] ["Red_clover" 0.005] ["Red_dead_nettle" 0.025] ["Rosebay_willowherb" 0.02] ["Selfheal" 0.08] ["Spear_thistle" 0.007] ["St_Johns_wort" 0.01] ["Tufted_vetch" 0.002] ["White_clover" 0.083] ["White_dead_nettle" 0.008]] 1470  "no info" 
49 "Hedgerow" 105 230.875 81.5 800 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 80  "no info" 
50 "Hedgerow" 105 97.462 78.462 1300 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 130  "no info" 
51 "Grass_improved" 115 186.083 61.441 167700 [["Dandelion" 0.006] ["White_clover" 0.47]] 3460  "no info" 
52 "Hedgerow" 105 146.8 72.8 2000 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 200  "no info" 
53 "Hedgerow" 105 175 67 8300 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 830  "no info" 
54 "Crop_Cereals" 75 202.412 74.706 1700 [["Crop_Cereals" 1]] 110  "no info" 
55 "Woodland" 35 154 58 34500 [["Birdsfoot_trefoil" 0.036] ["Blackthorn" 0.333] ["Bramble" 0.419] ["Bugle" 0.298] ["Burdock" 0.104] ["Buttercup" 0.724] ["Common_bluebell" 0.599] ["Common_knapweed" 0.008] ["Common_vetch" 0.038] ["Dandelion" 0.01] ["Dog_rose" 0.042] ["Field_scabious" 0.004] ["Foxglove" 0.02] ["Ground_ivy" 0.102] ["Hawthorn" 2.733] ["Hedge_woundwort" 0.038] ["Marsh_thistle" 0.023] ["Ragwort" 0.003] ["Red_clover" 0.005] ["Red_dead_nettle" 0.025] ["Rosebay_willowherb" 0.02] ["Selfheal" 0.08] ["Spear_thistle" 0.007] ["St_Johns_wort" 0.01] ["Tufted_vetch" 0.002] ["White_clover" 0.083] ["White_dead_nettle" 0.008]] 1540  "no info" 
56 "Hedgerow" 105 206 69 2000 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 200  "no info" 
57 "Woodland" 35 145 70 100 [["Birdsfoot_trefoil" 0.036] ["Blackthorn" 0.333] ["Bramble" 0.419] ["Bugle" 0.298] ["Burdock" 0.104] ["Buttercup" 0.724] ["Common_bluebell" 0.599] ["Common_knapweed" 0.008] ["Common_vetch" 0.038] ["Dandelion" 0.01] ["Dog_rose" 0.042] ["Field_scabious" 0.004] ["Foxglove" 0.02] ["Ground_ivy" 0.102] ["Hawthorn" 2.733] ["Hedge_woundwort" 0.038] ["Marsh_thistle" 0.023] ["Ragwort" 0.003] ["Red_clover" 0.005] ["Red_dead_nettle" 0.025] ["Rosebay_willowherb" 0.02] ["Selfheal" 0.08] ["Spear_thistle" 0.007] ["St_Johns_wort" 0.01] ["Tufted_vetch" 0.002] ["White_clover" 0.083] ["White_dead_nettle" 0.008]] 10  "no info" 
58 "Crop_Cereals" 75 215.571 67.971 3500 [["Crop_Cereals" 1]] 210  "no info" 
59 "Hedgerow" 105 135 59.684 1900 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 190  "no info" 
60 "Woodland" 35 218 64.333 600 [["Birdsfoot_trefoil" 0.036] ["Blackthorn" 0.333] ["Bramble" 0.419] ["Bugle" 0.298] ["Burdock" 0.104] ["Buttercup" 0.724] ["Common_bluebell" 0.599] ["Common_knapweed" 0.008] ["Common_vetch" 0.038] ["Dandelion" 0.01] ["Dog_rose" 0.042] ["Field_scabious" 0.004] ["Foxglove" 0.02] ["Ground_ivy" 0.102] ["Hawthorn" 2.733] ["Hedge_woundwort" 0.038] ["Marsh_thistle" 0.023] ["Ragwort" 0.003] ["Red_clover" 0.005] ["Red_dead_nettle" 0.025] ["Rosebay_willowherb" 0.02] ["Selfheal" 0.08] ["Spear_thistle" 0.007] ["St_Johns_wort" 0.01] ["Tufted_vetch" 0.002] ["White_clover" 0.083] ["White_dead_nettle" 0.008]] 60  "no info" 
61 "Grass_improved" 115 139.707 51.725 16700 [["Dandelion" 0.006] ["White_clover" 0.47]] 1040  "no info" 
62 "Hedgerow" 105 142.5 53.917 1200 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 120  "no info" 
63 "Crop_Cereals" 75 168.506 51.552 8700 [["Crop_Cereals" 1]] 520  "no info" 
64 "Hedgerow" 105 179.143 51 700 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 70  "no info" 
65 "Hedgerow" 105 153.5 50 800 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 80  "no info" 
66 "Hedgerow" 105 146.5 48.5 200 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 20  "no info" 
67 "Hedgerow" 105 160 48 300 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 30  "no info" 
68 "Hedgerow" 105 145 46 100 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 10  "no info" 
69 "Woodland" 35 197 43 5800 [["Birdsfoot_trefoil" 0.036] ["Blackthorn" 0.333] ["Bramble" 0.419] ["Bugle" 0.298] ["Burdock" 0.104] ["Buttercup" 0.724] ["Common_bluebell" 0.599] ["Common_knapweed" 0.008] ["Common_vetch" 0.038] ["Dandelion" 0.01] ["Dog_rose" 0.042] ["Field_scabious" 0.004] ["Foxglove" 0.02] ["Ground_ivy" 0.102] ["Hawthorn" 2.733] ["Hedge_woundwort" 0.038] ["Marsh_thistle" 0.023] ["Ragwort" 0.003] ["Red_clover" 0.005] ["Red_dead_nettle" 0.025] ["Rosebay_willowherb" 0.02] ["Selfheal" 0.08] ["Spear_thistle" 0.007] ["St_Johns_wort" 0.01] ["Tufted_vetch" 0.002] ["White_clover" 0.083] ["White_dead_nettle" 0.008]] 580  "no info" 
70 "Hedgerow" 105 212.75 39.083 1200 [["Average_Willow" 1.275] ["Birdsfoot_trefoil" 0.037] ["Blackthorn" 0.139] ["Bramble" 0.331] ["Bugle" 0.119] ["Burdock" 0.005] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Field_scabious" 0.008] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Ground_ivy" 0.063] ["Hawthorn" 0.696] ["Hedge_bindweed" 0] ["Hedge_woundwort" 0.043] ["Marsh_thistle" 0.004] ["Oilseed_rape" 0.052] ["Ragwort" 0.017] ["Red_clover" 0.046] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["Selfheal" 0.038] ["Spear_thistle" 0.008] ["St_Johns_wort" 0.002] ["Tufted_vetch" 0.046] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 120  "no info" 
71 "Grass_improved" 115 210.5 37.833 600 [["Dandelion" 0.006] ["White_clover" 0.47]] 60  "no info" 
72 "Woodland" 35 212.2 34.4 500 [["Birdsfoot_trefoil" 0.036] ["Blackthorn" 0.333] ["Bramble" 0.419] ["Bugle" 0.298] ["Burdock" 0.104] ["Buttercup" 0.724] ["Common_bluebell" 0.599] ["Common_knapweed" 0.008] ["Common_vetch" 0.038] ["Dandelion" 0.01] ["Dog_rose" 0.042] ["Field_scabious" 0.004] ["Foxglove" 0.02] ["Ground_ivy" 0.102] ["Hawthorn" 2.733] ["Hedge_woundwort" 0.038] ["Marsh_thistle" 0.023] ["Ragwort" 0.003] ["Red_clover" 0.005] ["Red_dead_nettle" 0.025] ["Rosebay_willowherb" 0.02] ["Selfheal" 0.08] ["Spear_thistle" 0.007] ["St_Johns_wort" 0.01] ["Tufted_vetch" 0.002] ["White_clover" 0.083] ["White_dead_nettle" 0.008]] 50  "no info" 
`

// ExampleFarmPlotTestFoodsources is netlogo/Example_Farm_PlotTest_Foodsources.txt, complete (39 lines, LF line
// endings as the original). The second map of the crop rotation experiment
// go/validation/experiments/step7_crop.xml: 36 flower patches of the example farm, 10 m per NetLogo patch.
const ExampleFarmPlotTestFoodsources = `10
36
id patchType patchColour xcor ycor size_sqm flowerSpeciesList perimeter_m info
73 "Crop_Cereals" 75 97.97 91.255 36500 [["Crop_Cereals" 1]] 940  "no info " 
107 "Crop_Cereals" 75 218.941 97.295 99000 [["Crop_Cereals" 1]] 2870  "no info " 
108 "Crop_Cereals" 75 125.948 141.856 105900 [["Crop_Cereals" 1]] 2080  "no info " 
111 "Woodland" 35 125.491 100.888 21400 [["Bramble" 0.419] ["Bugle" 0.298] ["Burdock" 0.104] ["Buttercup" 0.724] ["Common_bluebell" 0.599] ["Foxglove" 0.02] ["Hawthorn" 2.733] ["Marsh_thistle" 0.023] ["Rosebay_willowherb" 0.02]] 990  "no info " 
136 "Hedgerow" 105 145 107 8800 [["Average_Willow" 1.275] ["Bramble" 0.331] ["Buttercup" 0.381] ["Common_knapweed" 0.22] ["Foxglove" 0.065] ["Hawthorn" 0.696] ["Oilseed_rape" 0.052] ["Rosebay_willowherb" 0.127] ["Wild_teasel" 0.023]] 740  "no info " 
237 "Hedgerow" 105 107 121 12100 [["Average_Willow" 1.275] ["Bramble" 0.331] ["Buttercup" 0.381] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Dandelion" 0.11] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Hawthorn" 0.696] ["Oilseed_rape" 0.052] ["Rosebay_willowherb" 0.127] ["Wild_teasel" 0.023]] 1170  "no info " 
270 "Hedgerow" 105 184 86 2200 [["Average_Willow" 1.275] ["Bramble" 0.331] ["Foxglove" 0.065] ["Rosebay_willowherb" 0.127]] 170  "no info " 
330 "Hedgerow" 105 167.643 88.5 1400 [["Average_Willow" 1.275] ["Foxglove" 0.065]] 140  "no info " 
363 "Crop_Cereals" 75 202.412 74.706 1700 [["Crop_Cereals" 1]] 110  "no info " 
397 "Crop_Cereals" 75 124.333 95.333 300 [["Crop_Cereals" 1]] 30  "no info " 
499 "Woodland" 35 197 43 5800 [["Bramble" 0.419] ["Buttercup" 0.724] ["Common_bluebell" 0.599] ["Foxglove" 0.02] ["Hawthorn" 2.733]] 580  "no info " 
656 "Hedgerow" 105 146.8 72.8 2800 [["Average_Willow" 1.275] ["Bramble" 0.331] ["Common_knapweed" 0.22] ["Foxglove" 0.065] ["Rosebay_willowherb" 0.127]] 200  "no info " 
722 "Crop_Cereals" 75 168.506 51.552 8700 [["Crop_Cereals" 1]] 520  "no info " 
723 "Grassland" 135 191.922 86.398 33200 [["Birdsfoot_trefoil" 0.352] ["Buttercup" 2.161] ["Common_bluebell" 0.487] ["Dandelion" 0.121] ["Field_scabious" 0.017] ["Hawthorn" 0.617] ["Red_clover" 0.253] ["Tufted_vetch" 0.117] ["White_clover" 1.667]] 1300  "no info " 
746 "Grass_improved" 115 186.083 61.441 167700 [["White_clover" 0.47]] 3460  "no info " 
782 "Grassland" 135 111.035 113.783 11500 [["Birdsfoot_trefoil" 0.352] ["Buttercup" 2.161] ["Common_bluebell" 0.487] ["Dandelion" 0.121] ["Hawthorn" 0.617] ["White_clover" 1.667]] 630  "no info " 
804 "Crop_Cereals" 75 96 124 85700 [["Crop_Cereals" 0.942] ["Plot"] ["Birdsfoot_trefoil"  0.331] ["Common_mouse_ear"  0.062] ["Creeping_buttercup"  1.797] ["Dandelion"  0.001] ["Germander_speedwell"  0.01] ["Red_clover"  0.458] ["Red_dead_nettle"  0.002] ["Selfheal"  0.001] ["Smooth_hawksbeard"  0.006] ["White_clover"  0.101]] 2400  "5000" 
806 "Woodland" 35 154 58 34500 [["Blackthorn" 0.333] ["Bramble" 0.419] ["Bugle" 0.298] ["Burdock" 0.104] ["Buttercup" 0.724] ["Common_bluebell" 0.599] ["Foxglove" 0.02] ["Ground_ivy" 0.102] ["Hawthorn" 2.733] ["Marsh_thistle" 0.023] ["Rosebay_willowherb" 0.02]] 1540  "no info " 
931 "Hedgerow" 105 225 100 5400 [["Average_Willow" 1.275] ["Bramble" 0.331] ["Buttercup" 0.381] ["Common_knapweed" 0.22] ["Foxglove" 0.065] ["Hawthorn" 0.696] ["Rosebay_willowherb" 0.127]] 290  "no info " 
964 "Crop_Cereals" 75 215.571 67.971 3500 [["Crop_Cereals" 1]] 210  "no info " 
965 "Crop_Cereals" 75 128.85 83.948 156200 [["Crop_Cereals" 1]] 2500  "no info " 
966 "Scrub" 125 85.119 122.903 13400 [["Average_Willow" 1.208] ["Blackthorn" 1.292] ["Bramble" 0.211] ["Bugle" 0.509] ["Buttercup" 0.792] ["Common_bluebell" 0.256] ["Common_knapweed" 0.56] ["Dandelion" 0.113] ["Field_scabious" 0.023] ["Foxglove" 0.086] ["Hawthorn" 0.601] ["Hedge_woundwort" 0.287] ["Rosebay_willowherb" 0.043] ["Vipers_bugloss" 0.5] ["White_clover" 0.76] ["White_dead_nettle" 0.171] ["Wild_teasel" 0.037]] 500  "no info " 
999 "Crop_Cereals" 75 158.111 130.564 104600 [["Crop_Cereals" 1]] 2900  "no info " 
1001 "Woodland" 35 223.425 65.636 34100 [["Blackthorn" 0.333] ["Bramble" 0.419] ["Bugle" 0.298] ["Burdock" 0.104] ["Buttercup" 0.724] ["Common_bluebell" 0.599] ["Foxglove" 0.02] ["Ground_ivy" 0.102] ["Hawthorn" 2.733] ["Marsh_thistle" 0.023] ["Rosebay_willowherb" 0.02]] 1470  "no info " 
1027 "Hedgerow" 105 175 67 8300 [["Average_Willow" 1.275] ["Bramble" 0.331] ["Buttercup" 0.381] ["Common_knapweed" 0.22] ["Foxglove" 0.065] ["Hawthorn" 0.696] ["Rosebay_willowherb" 0.127] ["Wild_teasel" 0.023]] 830  "no info " 
1093 "Crop_Cereals" 75 196.446 115.052 28900 [["Crop_Cereals" 1]] 1080  "no info " 
1094 "Hedgerow" 105 221 70 10500 [["Average_Willow" 1.275] ["Bramble" 0.331] ["Buttercup" 0.381] ["Common_knapweed" 0.22] ["Dandelion" 0.11] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Hawthorn" 0.696] ["Oilseed_rape" 0.052] ["Rosebay_willowherb" 0.127] ["Wild_teasel" 0.023]] 650  "no info " 
1187 "Hedgerow" 105 179.143 51 1000 [["Average_Willow" 1.275] ["Foxglove" 0.065]] 70  "no info " 
1280 "Hedgerow" 105 181 141 100 [["Average_Willow" 1.275]] 10  "no info " 
1313 "Hedgerow" 105 129 132 5900 [["Average_Willow" 1.275] ["Bramble" 0.331] ["Buttercup" 0.381] ["Common_knapweed" 0.22] ["Foxglove" 0.065] ["Hawthorn" 0.696] ["Rosebay_willowherb" 0.127] ["Wild_teasel" 0.023]] 370  "no info " 
1347 "Woodland" 35 77.687 106.287 52700 [["Blackthorn" 0.333] ["Bramble" 0.419] ["Bugle" 0.298] ["Burdock" 0.104] ["Buttercup" 0.724] ["Common_bluebell" 0.599] ["Dog_rose" 0.042] ["Foxglove" 0.02] ["Ground_ivy" 0.102] ["Hawthorn" 2.733] ["Hedge_woundwort" 0.038] ["Marsh_thistle" 0.023] ["Rosebay_willowherb" 0.02]] 1440  "no info " 
1573 "Scrub" 125 202 121.5 3200 [["Average_Willow" 1.208] ["Bramble" 0.211] ["Buttercup" 0.792] ["Common_bluebell" 0.256] ["Common_knapweed" 0.56] ["Foxglove" 0.086]] 220  "no info " 
1640 "Grass_improved" 115 166.523 104.438 161300 [["White_clover" 0.47]] 3210  "no info " 
1641 "Hedgerow" 105 107.879 90.606 4600 [["Average_Willow" 1.275] ["Bramble" 0.331] ["Buttercup" 0.381] ["Common_knapweed" 0.22] ["Foxglove" 0.065] ["Hawthorn" 0.696] ["Rosebay_willowherb" 0.127]] 330  "no info " 
1701 "Hedgerow" 105 179 124 30000 [["Average_Willow" 1.275] ["Bramble" 0.331] ["Bugle" 0.119] ["Buttercup" 0.381] ["Comfrey" 0.042] ["Common_bluebell" 0.057] ["Common_knapweed" 0.22] ["Common_vetch" 0.132] ["Dandelion" 0.11] ["Dog_rose" 0.094] ["Foxglove" 0.065] ["Giant_bindweed" 0.047] ["Hawthorn" 0.696] ["Oilseed_rape" 0.052] ["Red_dead_nettle" 0.1] ["Rosebay_willowherb" 0.127] ["White_clover" 0.395] ["White_dead_nettle" 0.135] ["Wild_teasel" 0.023]] 1800  "no info " 
1767 "Hedgerow" 105 135 59.684 3400 [["Average_Willow" 1.275] ["Bramble" 0.331] ["Common_knapweed" 0.22] ["Foxglove" 0.065] ["Rosebay_willowherb" 0.127]] 190  "no info " 
`

// Flowerspecies is netlogo/_SYSTEM_Flowerspecies.csv, complete (80 lines). Names are stored as """Name""".
const Flowerspecies = `Flowerspecies,pollen_g/flower,nectar_ml/flower,proteinPollenProp,concentration_mol/l,startDay,stopDay,corollaDepth_mm,intFlowerTime_s
"""Alsike_clover""",0.0005,0.00046,0.2,0.7,181,242,10,0.6
"""American_willowherb""",0.00002,0.00035,0.2,1.34,152,243,3.5,0.6
"""Average_Willow""",0.01,0.00261,0.25,1.13,59,119,0,0.6
"""Bell_heather""",0.00073,0.0002,0.018285,1.17,121,334,5.5,0.6
"""Birdsfoot_trefoil""",0.00036,0.00035,0.31,0.96,120,303,9,0.6
"""Blackthorn""",0.00014,0.00009,0.23,0.77,59,150,0,0.6
"""Bramble""",0.00048,0.00682,0.21,0.5,120,272,0,0.6
"""Buckwheat""",0.00008,0.00011,0.11,1.33,162,222,0.5,0.6
"""Bugle""",0.00065,0.00044,0.14,0.67,120,211,10,0.6
"""Burdock""",0.00043,0.00229,0.11,0.88,181,272,3.9,0.6
"""Buttercup""",0.00077,0.0002,0.12,0.71,120,303,0,0.6
"""Charlock""",0.00026,0.00001,0.24,1.74,121,212,5.1,0.6
"""Chicory""",0.00086,0.00184,0.13,1.74,172,256,6.33,0.6
"""Comfrey""",0.00095,0.00432,0.17,0.98,120,211,17,0.6
"""Common_bluebell""",0.00188,0,0.36,0,90,180,0,0.6
"""Common_cats_ear""",0.43,0.00309,0.15,1.74,152,273,0,0.6
"""Common_field_speedwell""",0.00003,0.00049,0.08,0.1,1,365,0.5,0.6
"""Common_fumitory""",0,0.00003,0.24,1.83,121,304,7.5,0.6
"""Common_knapweed""",0.00239,0.00229,0.21,1.54,151,272,3,0.6
"""Common_mouse_ear""",0.01,0.00003,0.1,1.55,91,273,0,0.6
"""Common_vetch""",0.00038,0.00086,0.42,0.81,120,272,7,0.6
"""Creeping_buttercup""",0.00043,0.00035,0.1,0.85,121,304,0,0.6
"""Crop_Apple""",0.0017,0.0006,0.23,0.39,97,125,0.3,0.6
"""Crop_Cereals""",0,0,0,0,0,0,0,0.6
"""Crop_Field_beans""",0.00065,0.00086,0.2,1.28,153,182,19,0.6
"""Crop_Maize""",0.03,0,0.2,0,197,210,0,0.6
"""Crop_Oilseed_rape""",0.00024,0.00055,0.2,1.5,114,136,5,0.6
"""Cross_leaved_heath""",0.003,0.0002,0.31,1.17,152,273,6.5,0.6
"""Daisy""",0.22,0.00012,0.1,1.74,1,365,0,0.6
"""Dandelion""",0.00122,0.00154,0.09,1.07,1,364,1.2,0.6
"""Dog_rose""",0.00067,0,0.11,0,151,211,0,0.6
"""Field_pansy""",0.00097,0.00018,0.17,0.88,91,304,15,0.6
"""Field_scabious""",0.00488,0.0015,0.13,0.25,181,272,0,0.6
"""Field_woundwort""",0.0008,0.00087,0.14,1.04,91,334,5,0.6
"""Foxglove""",0.02,0.00163,0.21,0.82,151,272,7,0.6
"""Germander_speedwell""",0.00164,0.00033,0.08,0.1,60,212,0.5,0.6
"""Giant_bindweed""",0.00091,0.00995,0.26,0.66,181,272,0,0.6
"""Gorse""",0.00411,0.00002,0.22,1.74,60,273,13,0.6
"""Greater_birdsfoot_trefoil""",0.00098,0.00015,0.35,0.69,152,243,9,0.6
"""Greater_knapweed""",0.0024,0,0.26,0,181,272,13.6,0.6
"""Ground_ivy""",0.00046,0.00105,0.16,1.01,59,150,7,0.6
"""Hawthorn""",0.00016,0.00188,0.15,1.02,120,180,0,0.6
"""Heath_speedwell""",0.03,0.00003,0.2,1.37,121,304,3,0.6
"""Hedge_bindweed""",0.01,0.00715,0.13,1.4,182,273,0,0.6
"""Hedge_woundwort""",0.0008,0.00146,0.19,1.04,181,242,9,0.6
"""Ling""",0,0.0001,0.18,1.17,182,273,4,0.6
"""Marsh_thistle""",0.00505,0.00064,0.14,0.89,181,272,3,0.6
"""Meadow_buttercup""",0.0003,0.00026,0.17,0.94,121,304,0,0.6
"""Meadowsweet""",0.0077,0,0.19,1.74,152,273,0,0.6
"""Oilseed_rape""",0.00151,0.01,0.21,1.41,120,242,5,0.6
"""Ox-eye_daisy""",0.00122,0.00086,0.18,1.74,152,243,0,0.6
"""Phacelia""",0.00047,0.00028,0.3,1.64,152,243,7.5,0.6
"""Prickly_sow_thistle""",0.00045,0.00183,0.24,0.94,152,243,8.25,0.6
"""Ragwort""",0.00017,0,0.15,0,151,303,0,0.6
"""Red_campion""",0.0004,0.00335,0.15,2.35,91,273,16.1,0.6
"""Red_clover""",0.00018,0.00032,0.26,0.97,120,272,10,0.6
"""Red_dead_nettle""",0.00024,0.00282,0.14,1.37,59,303,7,0.6
"""Redshank""",0,0.00006,0.24,1.74,121,304,0.75,0.6
"""Ribwort_plantain""",0.00007,0,0.17,0.87,91,304,2.5,0.6
"""Rosebay_willowherb""",0.00612,0.00024,0.15,0.87,181,272,0,0.6
"""Salad_burnet""",0,0.00011,0,0,121,243,0,0.6
"""Scarlet_pimpernel""",0.00019,0,0.19,0,152,243,0,0.6
"""Selfheal""",0.00019,0.00039,0.2,0.87,151,272,8,0.6
"""Silverweed""",0.00053,0.00001,0.12,2.17,152,243,0,0.6
"""Smooth_hawksbeard""",0.00017,0.00001,0.1,1.34,152,273,3.8,0.6
"""Spear_thistle""",0.00254,0.00306,0.18,1.51,151,303,6.2,0.6
"""St_Johns_wort""",0.0005,0,0.13,0,151,272,0,0.6
"""Sunflower""",0.35,0.02,0.15,0.7,237,264,3.29,0.6
"""Tormentil""",0.00395,0.00004,0.16,2.17,152,273,0,0.6
"""Tufted_vetch""",0.00044,0.0007,0.12,0.61,151,242,6.8,0.6
"""Vipers_bugloss""",0.00071,0.00104,0.26,1.2,151,272,6.7,0.6
"""White_clover""",0.00014,0.00038,0.29,0.97,151,272,2,0.6
"""White_dead_nettle""",0.0007,0.00217,0.21,0.75,120,364,7.7,0.6
"""White_mustard""",0.00019,0.00061,0.24,0.9,152,243,4.2,0.6
"""Wild_carrot""",0.00002,0.00002,0.16,0.88,152,243,0,0.6
"""Wild_radish""",0.0007,0.00016,0.24,2.24,121,277,11.96,0.6
"""Wild_teasel""",0.01,0.00976,0.17,1.08,181,242,10,0.6
"""Yarrow""",0.00019,0.00003,0.08,1.79,152,243,2.4,0.6
"""Yellow_rattle""",0.01,0.00025,0.2,0.917323,121,243,13,0.6
`

// BumbleSpecies is netlogo/_SYSTEM_BumbleSpecies_UK_01.csv, complete (8 lines).
const BumbleSpecies = `species,name,emergingDay_mean,emergingDay_sd,nestHabitatsList,proboscis_min_mm,proboscis_max_mm,growthFactor,seasonStop,maxLifespanWorkers,batchsize,flightVelocity_m/s,searchLength_m,timeUnloading,specMax_cropVolume_myl,specMax_pollenPellets_g,minToMaxFactor,devAgeHatchingMin_d,devAgePupationMin_d,devAgeEmergingMin_d,devWeightEgg_mg,devWeightPupationMin_mg,devWeightPupationMax_mg,pollenToBodymassFactor,dev_Q_DeterminationWeight_mg,devAge_Q_PupationMin_d,devWeight_Q_PupationMin_mg,devWeight_Q_PupationMax_mg,devAge_Q_EmergingMin_d,dailyNestSiteChance
B_terrestris,B_terrestris,91,28,"[""Grassland*"" ""Garden"" ""Hedgerow"" ""Scrub"" ""Woodland"" ""Nestboxes""]",6.9,11.1,1.88,305,60,12,5,2500,165,173,0.15,2,5,12.9,23.8,1.5,62.4,249.5,1,0,17,590,980,32,0.2
B_pascuorum,B_pascuorum,136,14.5,"[""Grassland"" ""Garden"" ""Hedgerow"" ""Scrub"" ""Woodland"" ""Nestboxes""]",7.3,11.6,1.88,305,60,7,5,2500,165,173,0.15,2,5,12.9,23.8,1.5,48.5,187.6,1,0,17,261,461,32,0.2
B_lapidarius,B_lapidarius,143,20.5,"[""Grassland"" ""Garden"" ""Hedgerow"" ""Scrub"" ""Woodland"" ""Nestboxes""]",7,11.9,1.88,305,60,12,5,2500,165,173,0.15,2,5,12.9,23.8,1.5,90.7,159.8,1,0,17,300,850,32,0.2
B_hortorum,B_hortorum,120,8,"[""Grassland"" ""Garden"" ""Hedgerow"" ""Scrub"" ""Woodland"" ""Nestboxes""]",11.3,15.6,1.88,305,60,12,5,2500,165,173,0.15,2,5,12.9,23.8,1.5,71,271,1,0,17,371,751,32,0.2
B_pratorum,B_pratorum,110,30.5,"[""Garden"" ""Hedgerow"" ""Scrub"" ""Woodland"" ""Nestboxes""]",6.5,10.3,1.88,305,60,10,5,2500,165,173,0.15,2,5,12.9,23.8,1.5,51,231,1,0,17,271,491,32,0.2
B_hypnorum,B_hypnorum,100,10,"[""Garden"" ""Hedgerow"" ""Scrub"" ""Woodland""  ""Nestboxes""]",6.2,10.2,1.88,305,60,12,5,2500,165,173,0.15,2,5,12.9,19,1.5,112.4,218,1,0,17,334,410,32,0.2
Psithyrus,Psithyrus,120,15,"[""Grassland"" ""Garden"" ""Hedgerow"" ""Scrub"" ""Woodland"" ""Nestboxes""]",6.5,10.3,1.88,305,60,12,5,2500,165,173,0.15,2,5,12.9,23.8,1.5,140,170,1,0,17,590,980,32,0.2
`

// WeatherForagingHoursExample is netlogo/_SYSTEM_Weather_foragingHoursExample.csv, complete (366 lines): the daily
// foraging period [s] of NetLogo's Weather = "foragingHoursExample" [FPER-02] as a weather file of gobees-fast: day
// 1 to 365 = item 0 to 364 of the NetLogo list times 3600 s, rounded to whole seconds (finding 1, see
// trans/netlogo-findings.md#finding-1).
const WeatherForagingHoursExample = `Day,ForagingPeriod_s
1,0
2,11160
3,0
4,0
5,0
6,5400
7,0
8,360
9,0
10,0
11,6120
12,5760
13,0
14,0
15,0
16,0
17,0
18,0
19,0
20,0
21,5400
22,18000
23,0
24,11520
25,0
26,0
27,0
28,720
29,0
30,0
31,0
32,360
33,3240
34,21240
35,12600
36,24840
37,4680
38,27720
39,8280
40,16560
41,7920
42,0
43,1800
44,33120
45,0
46,28800
47,11520
48,14760
49,0
50,32400
51,32760
52,26280
53,20520
54,17640
55,0
56,43560
57,23400
58,28440
59,28440
60,39960
61,10080
62,0
63,10080
64,21600
65,20520
66,0
67,14400
68,36360
69,10440
70,36360
71,0
72,41040
73,22680
74,35640
75,15840
76,27000
77,28800
78,44280
79,31320
80,37080
81,13320
82,40680
83,47520
84,50400
85,15120
86,27720
87,29520
88,25920
89,33120
90,18000
91,47160
92,37800
93,12600
94,39960
95,48960
96,22320
97,30240
98,28080
99,30600
100,35280
101,23400
102,14760
103,38880
104,45000
105,54360
106,36360
107,15480
108,25200
109,33840
110,32040
111,27000
112,28080
113,23760
114,41040
115,43560
116,44640
117,42840
118,36360
119,52920
120,28080
121,47160
122,11880
123,59760
124,53280
125,64440
126,20520
127,720
128,10440
129,36000
130,52920
131,58320
132,56880
133,19080
134,20880
135,9000
136,21600
137,54720
138,4680
139,47160
140,40320
141,7200
142,46440
143,34920
144,7560
145,62280
146,20520
147,30600
148,47160
149,66600
150,6120
151,24120
152,49680
153,1800
154,2880
155,56520
156,17640
157,41040
158,42840
159,13680
160,42120
161,25560
162,76320
163,63720
164,6480
165,44280
166,56520
167,60840
168,60480
169,35640
170,12960
171,73440
172,46800
173,18360
174,2160
175,42120
176,7560
177,16920
178,50040
179,49680
180,5040
181,1080
182,66240
183,53280
184,46080
185,13320
186,48600
187,16920
188,1080
189,19800
190,14400
191,63000
192,6120
193,1080
194,53640
195,44640
196,41760
197,30600
198,16200
199,39960
200,57600
201,47520
202,49680
203,2520
204,25560
205,51480
206,12240
207,7920
208,20160
209,38160
210,12240
211,55800
212,56160
213,46080
214,54000
215,50400
216,21240
217,55800
218,32760
219,7200
220,3600
221,11520
222,33480
223,10800
224,11160
225,50400
226,36720
227,3600
228,34920
229,31680
230,13680
231,6840
232,42840
233,33480
234,23400
235,23760
236,30240
237,15480
238,25920
239,5400
240,41040
241,37440
242,48600
243,4320
244,21600
245,15840
246,48600
247,44640
248,28800
249,33480
250,21240
251,3240
252,24480
253,21240
254,32760
255,37800
256,21600
257,28440
258,8280
259,2880
260,0
261,28440
262,39600
263,4680
264,31320
265,23400
266,23760
267,27360
268,0
269,0
270,33840
271,25560
272,23040
273,14400
274,23760
275,0
276,9720
277,0
278,0
279,28080
280,0
281,31320
282,1080
283,7200
284,17280
285,6480
286,3240
287,0
288,0
289,25920
290,20880
291,23400
292,0
293,3960
294,0
295,0
296,2520
297,22680
298,4680
299,0
300,19800
301,5040
302,10080
303,0
304,0
305,0
306,14400
307,0
308,5040
309,18360
310,0
311,0
312,7560
313,0
314,1800
315,0
316,3600
317,0
318,0
319,8280
320,0
321,0
322,0
323,5040
324,2160
325,0
326,0
327,0
328,0
329,2880
330,0
331,0
332,3600
333,3240
334,0
335,0
336,0
337,0
338,0
339,0
340,8280
341,0
342,0
343,6840
344,5040
345,0
346,0
347,0
348,5400
349,0
350,0
351,0
352,3600
353,6840
354,0
355,0
356,12240
357,0
358,0
359,3600
360,0
361,0
362,0
363,0
364,0
365,0
`
